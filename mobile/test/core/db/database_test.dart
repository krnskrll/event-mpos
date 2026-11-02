import 'package:drift/drift.dart' show Value;
import 'package:drift/native.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mpos/core/db/database.dart';

void main() {
  late AppDatabase db;

  setUp(() => db = AppDatabase(NativeDatabase.memory()));
  tearDown(() => db.close());

  test('creates all tables', () async {
    final names = db.allTables.map((t) => t.actualTableName).toSet();
    expect(names, {
      'event_snapshot',
      'product',
      'member',
      'promotion',
      'pending_order',
      'receipt',
      'cart_draft',
      'number_block',
    });
  });

  test('stores and reads a pending order with millisecond timestamps', () async {
    final createdAt = DateTime.utc(2026, 10, 30, 9, 15, 30, 123);
    await db
        .into(db.pendingOrders)
        .insert(
          PendingOrdersCompanion.insert(
            idempotencyKey: 'k1',
            eventId: 'EV01',
            customerName: 'Walk-in',
            linesJson: '[]',
            status: 'PENDING',
            createdAt: createdAt,
          ),
        );

    final row = await db.select(db.pendingOrders).getSingle();
    expect(row.createdAt.toUtc(), createdAt);
    expect(row.orderNumber, isNull);
  });

  test('keeps what an order was charged (v3): subtotal, discount and promotion name', () async {
    await db
        .into(db.pendingOrders)
        .insert(
          PendingOrdersCompanion.insert(
            idempotencyKey: 'k2',
            eventId: 'EV01',
            customerName: 'Aisha',
            linesJson: '[]',
            status: 'PENDING',
            createdAt: DateTime(2026, 11, 2),
            subtotal: const Value(367000),
            discount: const Value(36700),
            promotionName: const Value('10% off everything'),
          ),
        );

    final row = await db.select(db.pendingOrders).getSingle();
    expect((row.subtotal, row.discount, row.promotionName), (367000.0, 36700.0, '10% off everything'));
  });
}
