import 'package:drift/drift.dart';
import 'package:drift_flutter/drift_flutter.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'tables.dart';

part 'database.g.dart';

@DriftDatabase(tables: [
  EventSnapshots,
  Products,
  Members,
  Promotions,
  PendingOrders,
  Receipts,
  CartDrafts,
  NumberBlocks,
])
class AppDatabase extends _$AppDatabase {
  AppDatabase([QueryExecutor? executor]) : super(executor ?? driftDatabase(name: 'mpos'));

  @override
  int get schemaVersion => 1;
}

final databaseProvider = Provider<AppDatabase>((ref) {
  final db = AppDatabase();
  ref.onDispose(db.close);
  return db;
});
