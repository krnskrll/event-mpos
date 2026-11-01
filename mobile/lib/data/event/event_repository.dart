import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:drift/drift.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/db/database.dart';
import '../../core/network/api_client.dart';
import '../../core/network/api_failure.dart';
import '../../domain/event.dart';

final eventRepositoryProvider = Provider((ref) => EventRepository(ref.watch(dioProvider), ref.watch(databaseProvider)));

class EventRepository {
  EventRepository(this._dio, this._db);

  final Dio _dio;
  final AppDatabase _db;

  Future<List<AssignedEvent>> assignedEvents() async {
    try {
      final res = await _dio.get<List>('api/me/events');
      return res.data!.map((e) => AssignedEvent.fromJson(e)).toList();
    } on DioException catch (e) {
      throw ApiFailure.from(e, 'Loading events');
    }
  }

  Future<String?> cachedVersion(String eventId) async {
    final query = _db.select(_db.eventSnapshots)..where((s) => s.eventId.equals(eventId));
    return (await query.getSingleOrNull())?.version;
  }

  /// Pull-to-refresh: compares the cheap snapshot hash first and downloads only when it moved.
  Future<void> refreshSnapshot(String eventId) async {
    try {
      final res = await _dio.get('api/me/events/$eventId/snapshot/hash');
      if (res.data['snapshotHash'] == await cachedVersion(eventId)) return;
    } on DioException catch (e) {
      throw ApiFailure.from(e, 'Refresh');
    }
    await syncSnapshot(eventId);
  }

  /// Downloads the event snapshot into the local DB. Sends the cached version so the server can
  /// answer "unchanged". A 409 (event no longer active) still completes: the cached data stays usable.
  Future<void> syncSnapshot(String eventId) async {
    final Map<String, dynamic> snapshot;
    try {
      final res = await _dio.get(
        'api/me/events/$eventId/snapshot',
        queryParameters: {'version': ?await cachedVersion(eventId)},
      );
      snapshot = res.data;
    } on DioException catch (e) {
      if (e.response?.statusCode == 409) return;
      throw ApiFailure.from(e, 'Sync', {403: 'You are not assigned to this event.'});
    }
    if (snapshot['changed'] == false) return;

    await _db.batch((b) {
      b.insert(
        _db.eventSnapshots,
        EventSnapshotsCompanion.insert(
          eventId: eventId,
          version: snapshot['version'],
          snapshotJson: jsonEncode(snapshot),
          syncedAt: DateTime.now(),
        ),
        mode: InsertMode.insertOrReplace,
      );
      b.deleteWhere(_db.products, (p) => p.eventId.equals(eventId));
      b.insertAll(_db.products, [for (final p in snapshot['products'] ?? []) _product(eventId, p)]);
      b.deleteWhere(_db.promotions, (p) => p.eventId.equals(eventId));
      b.insertAll(_db.promotions, [for (final p in snapshot['promotions'] ?? []) _promotion(eventId, p)]);
      // The member directory is global, not per event.
      b.deleteAll(_db.members);
      b.insertAll(_db.members, [
        for (final m in snapshot['members'] ?? [])
          MembersCompanion.insert(qrCode: m['qrCode'], name: m['name'], tier: m['tier'] ?? 'STANDARD'),
      ], mode: InsertMode.insertOrReplace);
    });
  }

  static ProductsCompanion _product(String eventId, Map<String, dynamic> p) => ProductsCompanion.insert(
    eventId: eventId,
    id: p['id'],
    sku: p['sku'],
    name: p['name'],
    brand: Value(p['brand']),
    price: (p['price'] as num? ?? 0).toDouble(),
    imageUrl: Value(p['imageUrl']),
    barcodes: (p['barcodes'] as List? ?? []).join('\n'),
    remainingStock: p['remainingStock'] ?? 0,
    variantsJson: Value(jsonEncode(p['variants'] ?? [])),
  );

  static PromotionsCompanion _promotion(String eventId, Map<String, dynamic> p) => PromotionsCompanion.insert(
    eventId: eventId,
    id: p['id'],
    name: p['name'],
    type: p['type'],
    percentOff: Value((p['percentOff'] as num?)?.toDouble()),
    amountOff: Value((p['amountOff'] as num?)?.toDouble()),
    scope: p['scope'] ?? 'ALL',
    productIds: (p['productIds'] as List? ?? []).join(','),
    minQty: Value(p['minQty']),
    minSubtotal: Value((p['minSubtotal'] as num?)?.toDouble()),
    startsAt: Value(p['startsAt']),
    endsAt: Value(p['endsAt']),
    priority: Value(p['priority'] ?? 0),
  );
}
