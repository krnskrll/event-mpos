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
      'event_snapshot', 'product', 'member', 'promotion',
      'pending_order', 'receipt', 'cart_draft', 'number_block',
    });
  });

  test('stores and reads a pending order with millisecond timestamps', () async {
    final createdAt = DateTime.utc(2026, 10, 30, 9, 15, 30, 123);
    await db.into(db.pendingOrders).insert(PendingOrdersCompanion.insert(
      idempotencyKey: 'k1',
      eventId: 'EV01',
      customerName: 'Walk-in',
      linesJson: '[]',
      status: 'PENDING',
      createdAt: createdAt,
    ));

    final row = await db.select(db.pendingOrders).getSingle();
    expect(row.createdAt.toUtc(), createdAt);
    expect(row.orderNumber, isNull);
  });
}
