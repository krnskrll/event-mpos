import 'package:dio/dio.dart';
import 'package:drift/drift.dart' show Value;
import 'package:drift/native.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mpos/core/db/database.dart';
import 'package:mpos/data/product/product_repository.dart';

import '../../support/fakes.dart';

void main() {
  late AppDatabase db;
  late FakeServer server;
  late ProductRepository repo;

  setUp(() async {
    db = AppDatabase(NativeDatabase.memory());
    for (final (id, variants) in [
      ('p1', '[{"id":"v1","name":"Rose","price":99000},{"id":"v2","name":"Coral","price":99000}]'),
      ('p2', '[]'),
    ]) {
      await db
          .into(db.products)
          .insert(
            ProductsCompanion.insert(
              eventId: 'EV01',
              id: id,
              sku: 'SKU-$id',
              name: 'Lipstick',
              price: 99000,
              barcodes: '111\n222',
              remainingStock: 0,
              variantsJson: Value(variants),
            ),
          );
    }
    server = FakeServer();
    repo = ProductRepository(Dio()..httpClientAdapter = server, db);
  });
  tearDown(() => db.close());

  Map<String, dynamic> levels({bool tracked = true}) => {
    'tracked': tracked,
    'placeId': tracked ? 'place-1' : null,
    'levels': [
      {'productId': 'p1', 'variantId': 'v1', 'available': 3},
      {'productId': 'p1', 'variantId': 'v2', 'available': 0},
      {'productId': 'p2', 'variantId': null, 'available': 8},
    ],
  };

  test('untracked: no limits, no badge', () async {
    server.routes['api/me/events/EV01/stock-levels'] = (200, levels(tracked: false));
    await repo.refreshStock('EV01');
    final p = (await repo.product('EV01', 'p1'))!;
    expect((p.available('v1'), p.shownStock, p.outOfStock), (null, null, false));
    expect(await repo.stockPlace('EV01'), isNull);
  });

  test('tracked: per variant, minus this device\'s unsent sales; the place is kept', () async {
    await db
        .into(db.pendingOrders)
        .insert(
          PendingOrdersCompanion.insert(
            idempotencyKey: 'k1',
            eventId: 'EV01',
            customerName: 'Walk-in',
            linesJson: '[{"productId":"p1","variantId":"v1","qty":1},{"productId":"p2","variantId":null,"qty":2}]',
            status: 'PENDING',
            createdAt: DateTime.now(),
          ),
        );
    server.routes['api/me/events/EV01/stock-levels'] = (200, levels());
    await repo.refreshStock('EV01');
    final lipstick = (await repo.product('EV01', 'p1'))!, other = (await repo.product('EV01', 'p2'))!;
    expect((lipstick.available('v1'), lipstick.available('v2'), lipstick.shownStock), (2, 0, 2));
    expect((other.available(null), other.lowStock), (6, true));
    expect(await repo.stockPlace('EV01'), 'place-1');

    await repo.adjustStock('EV01', 'p1', 'v1', -2);
    expect((await repo.product('EV01', 'p1'))!.outOfStock, isTrue);

    server.routes.remove('api/me/events/EV01/stock-levels'); // now answers 401, like a failed call
    await repo.refreshStock('EV01');
    expect((await repo.product('EV01', 'p2'))!.available(null), 6, reason: 'offline keeps the last numbers');
  });
}
