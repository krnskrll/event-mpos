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
  ],
)
class AppDatabase extends _$AppDatabase {
  // Shared so the background upload isolate and the app use one connection.
  AppDatabase([QueryExecutor? executor])
    : super(executor ?? driftDatabase(name: 'mpos', native: const DriftNativeOptions(shareAcrossIsolates: true)));

  @override
  int get schemaVersion => 4;

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
    },
  );
}

final databaseProvider = Provider<AppDatabase>((ref) {
  final db = AppDatabase();
  ref.onDispose(db.close);
  return db;
});
