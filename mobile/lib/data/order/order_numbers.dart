import 'package:dio/dio.dart';
import 'package:drift/drift.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../../core/db/database.dart';
import '../../core/network/api_client.dart';
import '../../core/preferences.dart';

final orderNumbersProvider = Provider(
  (ref) => OrderNumbers(ref.watch(databaseProvider), ref.watch(dioProvider), ref.watch(preferencesProvider)),
);

class NoOrderNumbers implements Exception {
  const NoOrderNumbers();

  @override
  String toString() =>
      'This device has run out of order numbers. Connect to the internet to get more, then confirm again.';
}

/// Final order numbers for offline selling, drawn from blocks the server reserved for this device.
class OrderNumbers {
  OrderNumbers(this._db, this._dio, this._prefs);

  static const _lowWater = 10;

  final AppDatabase _db;
  final Dio _dio;
  final SharedPreferences _prefs;
  Future<void>? _topUp;

  String get deviceId => deviceIdOf(_prefs);

  /// The next unused number and its business date, in one transaction. Throws [NoOrderNumbers].
  Future<({int number, String businessDate})> next(String eventId) => _db.transaction(() async {
    final block = (await _openBlocks(eventId)).firstOrNull;
    if (block == null) throw const NoOrderNumbers();
    await (_db.update(_db.numberBlocks)
          ..where((b) => b.eventId.equals(eventId) & b.rangeStart.equals(block.rangeStart)))
        .write(NumberBlocksCompanion(nextUnused: Value(block.nextUnused + 1)));
    return (number: block.nextUnused, businessDate: block.businessDate);
  });

  /// Fetches a fresh block when fewer than 10 numbers remain. Best effort: offline is fine.
  Future<void> topUp(String eventId) => _topUp ??= _fetchBlock(eventId).whenComplete(() => _topUp = null);

  Future<void> _fetchBlock(String eventId) async {
    if (await remaining(eventId) >= _lowWater) return;
    try {
      final block = (await _dio.post('api/me/events/$eventId/number-blocks', data: {'deviceId': deviceId})).data;
      await _db
          .into(_db.numberBlocks)
          .insert(
            NumberBlocksCompanion.insert(
              eventId: eventId,
              businessDate: block['businessDate'],
              rangeStart: block['rangeStart'],
              rangeEnd: block['rangeEnd'],
              nextUnused: block['rangeStart'],
            ),
          );
    } on DioException {
      return;
    }
  }

  /// Numbers this device can still hand out for [eventId], online or offline.
  Future<int> remaining(String eventId) async =>
      (await _openBlocks(eventId)).fold<int>(0, (sum, b) => sum + b.rangeEnd - b.nextUnused + 1);

  Future<List<NumberBlockRow>> _openBlocks(String eventId) =>
      (_db.select(_db.numberBlocks)
            ..where((b) => b.eventId.equals(eventId) & b.nextUnused.isSmallerOrEqual(b.rangeEnd))
            ..orderBy([(b) => OrderingTerm(expression: b.rangeStart)]))
          .get();
}
