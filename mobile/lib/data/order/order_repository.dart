import 'dart:async';
import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:drift/drift.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:uuid/uuid.dart';

import '../../core/config.dart';
import '../../core/db/database.dart';
import '../../core/network/api_client.dart';
import '../../domain/cart.dart';
import '../../domain/event.dart';
import '../../domain/order.dart';
import '../../domain/receipt.dart';
import '../product/product_repository.dart';
import '../receipt/receipt_service.dart';
import 'order_numbers.dart';
import 'order_uploader.dart';

final orderRepositoryProvider = Provider(
  (ref) => OrderRepository(
    ref.watch(databaseProvider),
    ref.watch(dioProvider),
    ref.watch(orderNumbersProvider),
    ref.watch(orderUploaderProvider),
    ref.watch(receiptServiceProvider),
    ref.watch(productRepositoryProvider),
  ),
);

class OrderRepository {
  OrderRepository(this._db, this._dio, this._numbers, this._uploader, this._receipts, this._products);

  static const _cashierNote = 'Please proceed to the cashier to complete payment.';

  final AppDatabase _db;
  final Dio _dio;
  final OrderNumbers _numbers;
  final OrderUploader _uploader;
  final ReceiptService _receipts;
  final ProductRepository _products;

  /// Confirms the sale on the device, online or offline: final number from this device's block,
  /// receipt stored and printed, upload queued as [staffId]. Returns the order key. Throws [NoOrderNumbers].
  Future<String> confirm(AssignedEvent event, Cart cart, PricedCart priced, {required String staffId}) async {
    final customer = cart.customer!;
    final (:number, :businessDate) = await _numbers.next(event.id);
    final padded = number.toString().padLeft(3, '0');
    final display = '#$padded';
    final reference = 'ORD-$businessDate-${event.code}-$padded';
    final key = const Uuid().v4();
    final now = DateTime.now();
    final lines = [
      for (final MapEntry(:key, value: qty) in cart.items.entries)
        {'productId': parseCartKey(key).productId, 'qty': qty, 'variantId': parseCartKey(key).variantId},
    ];

    await _db
        .into(_db.pendingOrders)
        .insert(
          PendingOrdersCompanion.insert(
            idempotencyKey: key,
            eventId: event.id,
            customerName: customer.name,
            customerQr: Value(customer.qrCode),
            customerTier: Value(customer.tier),
            linesJson: jsonEncode(lines),
            status: 'PENDING',
            createdAt: now,
            orderNumber: Value(number),
            orderNumberDisplay: Value(display),
            referenceId: Value(reference),
            total: Value(priced.total.toDouble()),
            subtotal: Value(priced.subtotal.toDouble()),
            discount: Value(priced.discount.toDouble()),
            promotionName: Value(priced.promotionName),
            staffId: Value(staffId),
          ),
        );
    for (final line in lines) {
      await _products.adjustStock(event.id, line['productId'] as String, -(line['qty'] as int));
    }
    _uploader.kick();
    unawaited(_numbers.topUp(event.id));

    // A printer problem never affects the order: the receipt is stored for reprint.
    await _receipts.storeAndPrint(
      key,
      event.id,
      ReceiptData(
        brand: brandName,
        eventName: event.name,
        orderNumberDisplay: display,
        referenceId: reference,
        customerName: customer.name,
        customerTier: customer.tier,
        items: [for (final i in priced.items) ReceiptLine(i.displayName, i.qty, i.lineTotal.toDouble())],
        subtotal: priced.subtotal.toDouble(),
        discount: priced.discount.toDouble(),
        promotionName: priced.promotionName,
        total: priced.total.toDouble(),
        cashierNote: _cashierNote,
        dateText: '${_two(now.day)}/${_two(now.month)} ${_two(now.hour)}:${_two(now.minute)}',
      ),
    );
    return key;
  }

  /// The event's orders, newest first.
  Stream<List<OrderRecord>> watchOrders(String eventId) {
    final query = _db.select(_db.pendingOrders)
      ..where((o) => o.eventId.equals(eventId))
      ..orderBy([(o) => OrderingTerm.desc(o.createdAt)]);
    return query.watch().map((rows) => rows.map(_toRecord).toList());
  }

  /// Cancels an uploaded order (online only) once a supervisor approved with their PIN, and returns
  /// its stock. Returns a message for the cashier when it can't.
  Future<String?> cancel(
    String key, {
    required String reason,
    required String supervisorId,
    required String pin,
  }) async {
    final order = await (_db.select(_db.pendingOrders)..where((o) => o.idempotencyKey.equals(key))).getSingleOrNull();
    if (order?.serverOrderId == null) return "This order hasn't synced yet — try again when it's confirmed.";
    try {
      await _dio.post(
        'api/me/orders/${order!.serverOrderId}/cancel',
        data: {'reason': reason, 'supervisorId': supervisorId, 'supervisorPin': pin},
      );
    } on DioException catch (e) {
      final data = e.response?.data;
      return switch (e.response?.statusCode) {
        null => "Can't reach the server. Cancel needs a connection.",
        409 => "The warehouse is already packing this order — it can't be cancelled here.",
        final code => (data is Map ? data['message'] as String? : null) ?? 'Cancel failed ($code)',
      };
    }
    await (_db.update(
      _db.pendingOrders,
    )..where((o) => o.idempotencyKey.equals(key))).write(const PendingOrdersCompanion(status: Value('CANCELLED')));
    await _receipts.markCancelled(key);
    for (final line in jsonDecode(order.linesJson)) {
      await _products.adjustStock(order.eventId, line['productId'], line['qty']);
    }
    return null;
  }

  static OrderRecord _toRecord(PendingOrderRow r) {
    final state = switch (r.status) {
      'CONFIRMED' => OrderState.confirmed,
      'FAILED' => OrderState.failed,
      'CANCELLED' => OrderState.cancelled,
      _ => OrderState.pending,
    };
    return OrderRecord(
      key: r.idempotencyKey,
      customerName: r.customerName,
      state: state,
      orderNumberDisplay: r.orderNumberDisplay,
      referenceId: r.referenceId,
      total: r.total,
      error: r.errorMessage,
      createdAt: r.createdAt,
      canCancel: state == OrderState.confirmed && r.serverOrderId != null,
      customerTier: r.customerTier,
      customerQr: r.customerQr,
      lines: [
        for (final l in jsonDecode(r.linesJson) as List)
          (productId: l['productId'] as String, variantId: l['variantId'] as String?, qty: l['qty'] as int),
      ],
      subtotal: r.subtotal,
      discount: r.discount,
      promotionName: r.promotionName,
    );
  }

  static String _two(int n) => n.toString().padLeft(2, '0');
}
