import 'package:dio/dio.dart';
import 'package:drift/drift.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/db/database.dart';
import '../../core/network/api_client.dart';
import '../../core/network/device_interceptor.dart';
import '../../core/printer/printer.dart';
import '../../domain/receipt.dart';

final receiptServiceProvider = Provider(
  (ref) => ReceiptService(ref.watch(databaseProvider), ref.watch(dioProvider), () => ref.read(printerProvider.future)),
);

/// Stores receipts and drives the printer. The receipt is saved before printing, so a printer
/// problem never loses it and never affects the order: the BA can reprint any time, offline.
class ReceiptService {
  ReceiptService(this._db, this._dio, this._printer);

  final AppDatabase _db;
  final Dio _dio;
  final Future<PrinterPort> Function() _printer;

  /// Returns an error message when printing failed, null when it printed.
  Future<String?> storeAndPrint(String orderKey, String eventId, ReceiptData data) async {
    final lines = ReceiptFormatter.format(data);
    await _db
        .into(_db.receipts)
        .insertOnConflictUpdate(
          ReceiptsCompanion.insert(
            orderKey: orderKey,
            eventId: eventId,
            orderNumberDisplay: data.orderNumberDisplay,
            receiptText: lines.join('\n'),
            printedOk: false,
            createdAt: DateTime.now(),
          ),
        );
    return _print(orderKey, lines);
  }

  Future<List<String>?> lines(String orderKey) async => (await _row(orderKey))?.receiptText.split('\n');

  /// Whether the last print of [orderKey]'s receipt succeeded; null when there is no receipt.
  Future<bool?> printed(String orderKey) async => (await _row(orderKey))?.printedOk;

  /// Prints the stored receipt again — no new order, no network needed.
  Future<String?> reprint(String orderKey) async {
    final row = await _row(orderKey);
    if (row == null) return 'No stored receipt for this order.';
    final error = await _print(orderKey, row.receiptText.split('\n'));
    if (row.backendOrderId != null) await _logPrint(row.backendOrderId!);
    return error;
  }

  /// After upload: remember the server order (reprints are logged against it) and log the first print
  /// as [staffId], the cashier who sold it.
  Future<void> linkServerOrder(String orderKey, String serverOrderId, {String? staffId}) async {
    await _update(orderKey, ReceiptsCompanion(backendOrderId: Value(serverOrderId)));
    await _logPrint(serverOrderId, staffId);
  }

  /// Stamps the stored receipt CANCELLED, so the screen and every reprint show it.
  Future<void> markCancelled(String orderKey) async {
    final row = await _row(orderKey);
    if (row == null || row.receiptText.startsWith(ReceiptFormatter.cancelled)) return;
    await _update(orderKey, ReceiptsCompanion(receiptText: Value('${ReceiptFormatter.cancelled}\n${row.receiptText}')));
  }

  Future<String?> _print(String orderKey, List<String> lines) async {
    final printer = await _printer();
    var status = await printer.status();
    if (status == PrinterStatus.ready) status = await printer.print(lines);
    await _update(orderKey, ReceiptsCompanion(printedOk: Value(status == PrinterStatus.ready)));
    return status == PrinterStatus.ready ? null : 'Printer ${status.label}';
  }

  /// Best effort: the server audit counts prints and reprints.
  Future<void> _logPrint(String serverOrderId, [String? staffId]) => _dio
      .post('api/me/orders/$serverOrderId/receipts', options: Options(headers: {staffHeader: ?staffId}))
      .then((_) {}, onError: (_) {});

  Future<ReceiptRow?> _row(String orderKey) =>
      (_db.select(_db.receipts)..where((r) => r.orderKey.equals(orderKey))).getSingleOrNull();

  Future<void> _update(String orderKey, ReceiptsCompanion values) =>
      (_db.update(_db.receipts)..where((r) => r.orderKey.equals(orderKey))).write(values);
}
