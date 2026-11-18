import 'package:drift/drift.dart';
import 'package:drift_flutter/drift_flutter.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'tables.dart';

part 'database.g.dart';

@DriftDatabase(
  tables: [
    EventSnapshots,
    Products,
    Members,
    Promotions,
    PendingOrders,
    Receipts,
    CartDrafts,
    NumberBlocks,
    StaffMembers,
    EventStocks,
  ],
)
class AppDatabase extends _$AppDatabase {
  // Shared so the background upload isolate and the app use one connection.
  AppDatabase([QueryExecutor? executor])
    : super(executor ?? driftDatabase(name: 'mpos', native: const DriftNativeOptions(shareAcrossIsolates: true)));

  @override
  int get schemaVersion => 5;

  @override
  MigrationStrategy get migration => MigrationStrategy(
    onUpgrade: (m, from, to) async {
      if (from < 2) await m.addColumn(pendingOrders, pendingOrders.claimedAt);
      if (from < 3) {
        await m.addColumn(pendingOrders, pendingOrders.subtotal);
        await m.addColumn(pendingOrders, pendingOrders.discount);
        await m.addColumn(pendingOrders, pendingOrders.promotionName);
      }
      if (from < 4) {
        await m.addColumn(pendingOrders, pendingOrders.staffId);
        await m.createTable(staffMembers);
      }
      if (from < 5) {
        await m.addColumn(products, products.stockTracked);
        await m.addColumn(products, products.variantStockJson);
        await m.addColumn(pendingOrders, pendingOrders.stockPlaceId);
        await m.createTable(eventStocks);
      }
    },
  );

  /// Orders sold here that the server hasn't confirmed yet (6e: reported as X-Pending-Orders).
  Future<int> unsentOrderCount() async {
    final n = pendingOrders.idempotencyKey.count();
    final q = selectOnly(pendingOrders)
      ..addColumns([n])
      ..where(pendingOrders.status.isIn(const ['PENDING', 'UPLOADING']));
    return (await q.getSingle()).read(n) ?? 0;
  }
}

final databaseProvider = Provider<AppDatabase>((ref) {
  final db = AppDatabase();
  ref.onDispose(db.close);
  return db;
});
