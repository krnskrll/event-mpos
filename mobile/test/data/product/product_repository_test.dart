import 'package:dio/dio.dart';
import 'package:drift/drift.dart' show Value;
import 'package:drift/native.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mpos/core/db/database.dart';
import 'package:mpos/data/product/product_repository.dart';

import '../../support/fakes.dart';

void main() {
  test('maps stored rows and applies live stock; offline keeps the last levels', () async {
    final db = AppDatabase(NativeDatabase.memory());
    addTearDown(db.close);
    await db
        .into(db.products)
        .insert(
          ProductsCompanion.insert(
            eventId: 'EV01',
            id: 'p1',
            sku: 'SKU-1',
            name: 'Lipstick',
            price: 99000,
            barcodes: '111\n222',
            remainingStock: 5,
            variantsJson: const Value('[{"id":"v1","name":"Rose","price":99000}]'),
          ),
        );
    final server = FakeServer()
      ..routes['api/me/events/EV01/stock'] = (
        200,
        [
          {'productId': 'p1', 'remaining': 0},
        ],
      );
    final repo = ProductRepository(Dio()..httpClientAdapter = server, db);

    final product = (await repo.product('EV01', 'p1'))!;
    expect(product.barcodes, ['111', '222']);
    expect((product.variants.single.name, product.lowStock), ('Rose', true));

    await repo.refreshStock('EV01');
    expect((await repo.product('EV01', 'p1'))!.outOfStock, isTrue);

    server.routes.remove('api/me/events/EV01/stock'); // now answers 401, like a failed call
    await repo.refreshStock('EV01');
    expect((await repo.product('EV01', 'p1'))!.remainingStock, 0);
  });
}
