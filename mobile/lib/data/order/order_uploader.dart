import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'package:dio/dio.dart';
import 'package:drift/drift.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:workmanager/workmanager.dart';

import '../../core/db/database.dart';
import '../../core/format/money.dart';
import '../../core/network/api_client.dart';
import '../../core/notifications.dart';
import '../receipt/receipt_service.dart';
import 'order_numbers.dart';

final orderUploaderProvider = Provider(
  (ref) => OrderUploader(
    ref.watch(dioProvider),
    ref.watch(databaseProvider),
    ref.watch(receiptServiceProvider),
    ref.watch(orderNumbersProvider),
  ),
);

const uploadTask = 'submit-orders';

/// Uploads orders already confirmed on the device, oldest first. The idempotency key makes every
/// retry duplicate-free. Outcome per order:
///  - success              -> CONFIRMED (server id kept for cancel and receipt logging)
///  - network error / 5xx  -> stays PENDING, retried later
///  - other 4xx            -> FAILED with the reason (needs the BA)
class OrderUploader {
  OrderUploader(this._dio, this._db, this._receipts, this._numbers, {this.notify = notifyOrder});

  /// An upload claim older than this is considered abandoned (the app was killed mid-upload).
  static const _claimTimeout = Duration(minutes: 2);

  final Dio _dio;
  final AppDatabase _db;
  final ReceiptService _receipts;
  final OrderNumbers _numbers;
  final Future<void> Function(String key, String title, String body) notify;

  /// Uploads now, and on Android also queues a WorkManager job that retries in the background.
  void kick() {
    unawaited(drainPending());
    if (Platform.isAndroid) {
      Workmanager().registerOneOffTask(
        uploadTask,
        uploadTask,
        constraints: Constraints(networkType: NetworkType.connected),
        existingWorkPolicy: ExistingWorkPolicy.append,
        backoffPolicy: BackoffPolicy.exponential,
        backoffPolicyDelay: const Duration(seconds: 10),
      );
    }
  }

  /// Returns true when at least one order must be retried later.
  Future<bool> drainPending() async {
    final stale = DateTime.now().subtract(_claimTimeout);
    final waiting =
        await (_db.select(_db.pendingOrders)
              ..where((o) => _claimable(o, stale))
              ..orderBy([(o) => OrderingTerm(expression: o.createdAt)]))
            .get();
    var retry = false;
    for (final order in waiting) {
      if (await _claim(order.idempotencyKey, stale)) retry |= !await _upload(order);
    }
    return retry;
  }

  Expression<bool> _claimable($PendingOrdersTable o, DateTime stale) =>
      o.status.equals('PENDING') | (o.status.equals('UPLOADING') & o.claimedAt.isSmallerThanValue(stale));

  /// Marks the order UPLOADING unless another isolate already holds it. The background job and the
  /// app share the database, not memory, so the claim lives in the row.
  Future<bool> _claim(String key, DateTime stale) async =>
      await (_db.update(_db.pendingOrders)..where((o) => o.idempotencyKey.equals(key) & _claimable(o, stale))).write(
        PendingOrdersCompanion(status: const Value('UPLOADING'), claimedAt: Value(DateTime.now())),
      ) ==
      1;

  /// Returns false when the order must be retried.
  Future<bool> _upload(PendingOrderRow order) async {
    try {
      final res = await _dio.post(
        'api/me/events/${order.eventId}/orders',
        data: {
          'idempotencyKey': order.idempotencyKey,
          'customer': {'name': order.customerName, 'qrCode': order.customerQr, 'tier': order.customerTier},
          'lines': jsonDecode(order.linesJson),
          'deviceId': _numbers.deviceId,
          'orderNumber': order.orderNumber,
          'clientCreatedAt': order.createdAt.toUtc().toIso8601String(),
        },
      );
      final o = res.data;
      await _set(
        order.idempotencyKey,
        PendingOrdersCompanion(
          status: const Value('CONFIRMED'),
          orderNumberDisplay: Value(o['orderNumberDisplay']),
          referenceId: Value(o['referenceId']),
          total: Value((o['total'] as num).toDouble()),
          serverOrderId: Value(o['id']),
          errorMessage: const Value(null),
        ),
      );
      await _receipts.linkServerOrder(order.idempotencyKey, o['id']);
      await notify(order.idempotencyKey, 'Order confirmed', '${o['orderNumberDisplay']} · ${rupiah(o['total'])}');
      return true;
    } on DioException catch (e) {
      final status = e.response?.statusCode;
      if (status == null || status >= 500) {
        await _set(order.idempotencyKey, const PendingOrdersCompanion(status: Value('PENDING')));
        return false;
      }
      final reason = 'Rejected ($status)';
      await _set(
        order.idempotencyKey,
        PendingOrdersCompanion(status: const Value('FAILED'), errorMessage: Value(reason)),
      );
      await notify(order.idempotencyKey, 'Order failed — ${order.customerName}', reason);
      return true;
    }
  }

  Future<void> _set(String key, PendingOrdersCompanion values) =>
      (_db.update(_db.pendingOrders)..where((o) => o.idempotencyKey.equals(key))).write(values);
}
