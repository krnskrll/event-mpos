import 'package:dio/dio.dart';
import 'package:drift/drift.dart' show Value;
import 'package:drift/native.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mpos/core/db/database.dart';
import 'package:mpos/core/network/device_interceptor.dart';
import 'package:mpos/core/printer/printer.dart';
import 'package:mpos/data/order/order_uploader.dart';
import 'package:mpos/data/receipt/receipt_service.dart';

import '../../support/fakes.dart';

const _path = 'api/me/events/EV01/orders';

void main() {
  late AppDatabase db;
  late FakeServer server;
  late OrderUploader uploader;

  setUp(() async {
    db = AppDatabase(NativeDatabase.memory());
    server = FakeServer();
    // Sari is signed in now; the queued order was sold by Budi.
    final dio = Dio()
      ..httpClientAdapter = server
      ..interceptors.add(DeviceInterceptor(await fakeDevice(), () => sari.id));
    uploader = OrderUploader(
      dio,
      db,
      ReceiptService(db, dio, () async => const FakePrinter()),
      notify: (_, _, _) async {},
    );
    await db
        .into(db.pendingOrders)
        .insert(
          PendingOrdersCompanion.insert(
            idempotencyKey: 'k1',
            eventId: 'EV01',
            customerName: 'Aisha',
            linesJson: '[{"productId":"p1","qty":1}]',
            status: 'PENDING',
            createdAt: DateTime(2026, 10, 31, 10),
            orderNumber: const Value(7),
            staffId: Value(budi.id),
            stockPlaceId: const Value('place-1'),
          ),
        );
  });
  tearDown(() => db.close());

  Future<PendingOrderRow> order() => db.select(db.pendingOrders).getSingle();

  test('network trouble keeps the order queued; a rejection fails it; success confirms it', () async {
    server.routes[_path] = (503, {});
    expect(await uploader.drainPending(), isTrue);
    expect((await order()).status, 'PENDING');

    server.routes[_path] = (200, {'id': 's1', 'orderNumberDisplay': '#007', 'referenceId': 'ORD-1', 'total': 99000});
    expect(await uploader.drainPending(), isFalse);
    expect(((await order()).status, (await order()).serverOrderId), ('CONFIRMED', 's1'));
    expect(server.requests.first.data['lines'], [
      {'productId': 'p1', 'qty': 1},
    ]);
    // 6c: the sale carries the place it sold from, even when it uploads late.
    expect((server.requests.first.data['stockPlaceId'], server.requests.first.data['stockKnown']), ('place-1', true));

    await (db.update(db.pendingOrders)).write(const PendingOrdersCompanion(status: Value('PENDING')));
    server.routes[_path] = (409, {});
    expect(await uploader.drainPending(), isFalse);
    expect(((await order()).status, (await order()).errorMessage), ('FAILED', 'Rejected (409)'));

    await (db.update(db.pendingOrders)).write(const PendingOrdersCompanion(status: Value('PENDING')));
    server.routes[_path] = (403, {'message': 'Unknown staff member'});
    await uploader.drainPending();
    expect((await order()).errorMessage, 'Unknown staff member');
  });

  test('uploads as the cashier who sold the order, with the device token', () async {
    server.routes[_path] = (200, {'id': 's1', 'orderNumberDisplay': '#007', 'referenceId': 'ORD-1', 'total': 99000});
    await uploader.drainPending();

    final upload = server.requests.firstWhere((r) => r.path == _path);
    expect((upload.headers['Authorization'], upload.headers[staffHeader]), ('Device t1', budi.id));
    expect(upload.data.containsKey('deviceId'), isFalse);
    final receiptLog = server.requests.firstWhere((r) => r.path == 'api/me/orders/s1/receipts');
    expect(receiptLog.headers[staffHeader], budi.id);
  });

  test('6e: each upload reports the orders still waiting after it', () async {
    await db
        .into(db.pendingOrders)
        .insert(
          PendingOrdersCompanion.insert(
            idempotencyKey: 'k2',
            eventId: 'EV01',
            customerName: 'Rina',
            linesJson: '[]',
            status: 'PENDING',
            createdAt: DateTime(2026, 10, 31, 11),
          ),
        );
    server.routes[_path] = (200, {'id': 's1', 'orderNumberDisplay': '#007', 'referenceId': 'ORD-1', 'total': 99000});
    await uploader.drainPending();

    final uploads = server.requests.where((r) => r.path == _path).map((r) => r.headers[pendingHeader]);
    expect(uploads, ['1', '0']);
    expect(await db.unsentOrderCount(), 0);
  });

  test('a removed device (401) keeps the order queued without retrying', () async {
    server.routes[_path] = (401, {});
    expect(await uploader.drainPending(), isFalse);
    expect((await order()).status, 'PENDING');
  });

  test('two drains at once upload the order only once', () async {
    server
      ..delay = const Duration(milliseconds: 50)
      ..routes[_path] = (200, {'id': 's1', 'orderNumberDisplay': '#007', 'referenceId': 'ORD-1', 'total': 99000});

    await Future.wait([uploader.drainPending(), uploader.drainPending()]);
    expect(server.requests.where((r) => r.path == _path), hasLength(1));
  });
}
