import 'package:dio/dio.dart';
import 'package:drift/native.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mpos/core/db/database.dart';
import 'package:mpos/core/network/api_failure.dart';
import 'package:mpos/data/event/event_repository.dart';

import '../../support/fakes.dart';

const _path = 'api/me/events/EV01/snapshot';

const _snapshot = {
  'version': 'v1',
  'products': [
    {
      'id': 'p1',
      'sku': 'SKU-1',
      'name': 'Serum',
      'price': 129000,
      'barcodes': ['111', '222'],
      'variants': [],
      'remainingStock': 23,
    },
  ],
  'members': [
    {'qrCode': 'MBR-1', 'name': 'Aisha', 'tier': 'PLATINUM'},
  ],
  'promotions': [
    {
      'id': 'pr1',
      'name': '10% off',
      'type': 'PERCENTAGE',
      'percentOff': 10,
      'scope': 'ALL',
      'productIds': [],
      'priority': 1,
    },
  ],
};

void main() {
  late FakeServer server;
  late AppDatabase db;
  late EventRepository repo;

  setUp(() {
    server = FakeServer()..routes[_path] = (200, _snapshot);
    db = AppDatabase(NativeDatabase.memory());
    repo = EventRepository(Dio()..httpClientAdapter = server, db);
  });
  tearDown(() => db.close());

  test('stores the snapshot, then sends its version and keeps it when unchanged', () async {
    await repo.syncSnapshot('EV01');
    final product = await db.select(db.products).getSingle();
    expect((product.price, product.barcodes), (129000.0, '111\n222'));
    expect(await db.select(db.members).get(), hasLength(1));
    expect(await db.select(db.promotions).get(), hasLength(1));

    server.routes[_path] = (200, {'changed': false, 'version': 'v1'});
    await repo.syncSnapshot('EV01');
    expect(server.requests.last.queryParameters, {'version': 'v1'});
    expect(await db.select(db.products).get(), hasLength(1));
  });

  test('403 explains the assignment; 409 (event closed) still opens the cached session', () async {
    server.routes[_path] = (403, {});
    await expectLater(
      repo.syncSnapshot('EV01'),
      throwsA(isA<ApiFailure>().having((e) => e.message, 'message', 'You are not assigned to this event.')),
    );

    server.routes[_path] = (409, {'message': 'Event is closed'});
    await repo.syncSnapshot('EV01');
  });
}
