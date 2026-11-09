import 'package:dio/dio.dart';
import 'package:drift/drift.dart' show Value;
import 'package:drift/native.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mpos/core/db/database.dart';
import 'package:mpos/core/network/device_interceptor.dart';
import 'package:mpos/core/printer/printer.dart';
import 'package:mpos/data/order/order_numbers.dart';
import 'package:mpos/data/order/order_repository.dart';
import 'package:mpos/data/order/order_uploader.dart';
import 'package:mpos/data/product/product_repository.dart';
import 'package:mpos/data/receipt/receipt_service.dart';

import '../../support/fakes.dart';

const _path = 'api/me/orders/s1/cancel';

void main() {
  late AppDatabase db;
  late FakeServer server;
  late OrderRepository repo;

  setUp(() async {
    db = AppDatabase(NativeDatabase.memory());
    server = FakeServer();
    final dio = Dio()
      ..httpClientAdapter = server
      ..interceptors.add(DeviceInterceptor(await fakeDevice(), () => budi.id));
    final receipts = ReceiptService(db, dio, () async => const FakePrinter());
    repo = OrderRepository(
      db,
      dio,
      OrderNumbers(db, dio),
      OrderUploader(dio, db, receipts, notify: (_, _, _) async {}),
      receipts,
      ProductRepository(dio, db),
    );
    await db
        .into(db.pendingOrders)
        .insert(
          PendingOrdersCompanion.insert(
            idempotencyKey: 'k1',
            eventId: 'EV01',
            customerName: 'Aisha',
            linesJson: '[]',
            status: 'CONFIRMED',
            createdAt: DateTime(2026, 11, 9, 10),
            serverOrderId: const Value('s1'),
          ),
        );
  });
  tearDown(() => db.close());

  Future<String> status() async => (await db.select(db.pendingOrders).getSingle()).status;

  test('a cancel carries the supervisor and their PIN; server refusals are shown as sent', () async {
    server.routes[_path] = (403, {'message': 'Wrong supervisor PIN'});
    expect(await repo.cancel('k1', reason: 'Out of stock', supervisorId: sari.id, pin: '1111'), 'Wrong supervisor PIN');
    expect(await status(), 'CONFIRMED');
    final sent = server.requests.last;
    expect(sent.data, {'reason': 'Out of stock', 'supervisorId': sari.id, 'supervisorPin': '1111'});
    expect(sent.headers[staffHeader], budi.id); // the cashier asking; the supervisor approves

    server.routes[_path] = (409, {});
    expect(await repo.cancel('k1', reason: 'Out of stock', supervisorId: sari.id, pin: '482913'), contains('packing'));

    server.routes[_path] = (200, {'id': 's1', 'status': 'CANCELLED'});
    expect(await repo.cancel('k1', reason: 'Out of stock', supervisorId: sari.id, pin: '482913'), isNull);
    expect(await status(), 'CANCELLED');
  });
}
