import 'dart:convert';

import 'package:drift/drift.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:uuid/uuid.dart';

import '../../core/db/database.dart';
import '../../domain/cart.dart';

final draftRepositoryProvider = Provider((ref) => DraftRepository(ref.watch(databaseProvider)));

typedef CartDraft = ({String id, String label, Cart cart, DateTime createdAt});

/// Parked carts per event, plus one autosaved copy of the live cart that survives the app being killed.
class DraftRepository {
  DraftRepository(this._db);

  final AppDatabase _db;

  static String _liveId(String eventId) => 'live:$eventId';

  Stream<List<CartDraft>> watchDrafts(String eventId) {
    final query = _db.select(_db.cartDrafts)
      ..where((d) => d.eventId.equals(eventId) & d.live.equals(false))
      ..orderBy([(d) => OrderingTerm.desc(d.createdAt)]);
    return query.watch().map(
      (rows) => [for (final r in rows) (id: r.id, label: r.label, cart: _cart(r), createdAt: r.createdAt)],
    );
  }

  Future<void> saveDraft(String eventId, String label, Cart cart) =>
      _write(const Uuid().v4(), eventId, label.trim().isEmpty ? 'Draft' : label.trim(), cart, live: false);

  /// Loads a draft and removes it from the list.
  Future<Cart?> take(String id) async {
    final row = await (_db.select(_db.cartDrafts)..where((d) => d.id.equals(id))).getSingleOrNull();
    await delete(id);
    return row == null ? null : _cart(row);
  }

  Future<void> delete(String id) => (_db.delete(_db.cartDrafts)..where((d) => d.id.equals(id))).go();

  Future<Cart?> loadLive(String eventId) async {
    final row = await (_db.select(_db.cartDrafts)..where((d) => d.id.equals(_liveId(eventId)))).getSingleOrNull();
    return row == null ? null : _cart(row);
  }

  Future<void> saveLive(String eventId, Cart cart) =>
      cart.isEmpty ? delete(_liveId(eventId)) : _write(_liveId(eventId), eventId, '', cart, live: true);

  /// Drops every draft and the autosaved cart of an event (switched away from, or closed).
  Future<void> clearEvent(String eventId) => (_db.delete(_db.cartDrafts)..where((d) => d.eventId.equals(eventId))).go();

  Future<void> _write(String id, String eventId, String label, Cart cart, {required bool live}) {
    final json = cart.toJson();
    return _db
        .into(_db.cartDrafts)
        .insertOnConflictUpdate(
          CartDraftsCompanion.insert(
            id: id,
            eventId: eventId,
            label: label,
            customerJson: Value(json['customer'] == null ? null : jsonEncode(json['customer'])),
            linesJson: jsonEncode(json['items']),
            createdAt: DateTime.now(),
            live: Value(live),
          ),
        );
  }

  static Cart _cart(CartDraftRow r) => Cart.fromJson({
    'items': jsonDecode(r.linesJson),
    'customer': r.customerJson == null ? null : jsonDecode(r.customerJson!),
  });
}
