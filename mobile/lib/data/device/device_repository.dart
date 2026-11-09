import 'dart:async';
import 'dart:io';

import 'package:dio/dio.dart';
import 'package:drift/drift.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/db/database.dart';
import '../../core/network/api_client.dart';
import '../../core/network/api_failure.dart';
import '../../core/secure/device_storage.dart';
import '../../domain/device.dart';
import '../../domain/staff.dart';
import '../event/event_repository.dart';
import '../session/session.dart';

final deviceRepositoryProvider = Provider(
  (ref) => DeviceRepository(
    ref.watch(dioProvider),
    ref.watch(databaseProvider),
    ref.watch(deviceStorageProvider),
    ref.watch(eventRepositoryProvider),
  ),
);

final staffProvider = StreamProvider((ref) => ref.watch(deviceRepositoryProvider).watchStaff());

/// The location's supervisors: the only staff who can approve a cancel.
final supervisorsProvider = Provider(
  (ref) => [
    for (final s in ref.watch(staffProvider).value ?? const <StaffMember>[])
      if (s.isSupervisor) s,
  ],
);

/// Keeps the pairing fresh while the app runs: device, staff and snapshot at start, on resume and
/// every minute. Best effort; a removed device (401) is handled by the interceptor.
final deviceSyncProvider = Provider<void>((ref) {
  if (ref.watch(deviceProvider.select((d) => d?.id)) == null) return;
  final repo = ref.read(deviceRepositoryProvider);
  Future<void> refresh() => repo.refresh().catchError((_) {});
  unawaited(refresh());
  final timer = Timer.periodic(const Duration(seconds: 60), (_) => refresh());
  final lifecycle = AppLifecycleListener(onResume: refresh);
  ref.onDispose(() {
    timer.cancel();
    lifecycle.dispose();
  });
});

class DeviceRepository {
  DeviceRepository(this._dio, this._db, this._storage, this._events);

  final Dio _dio;
  final AppDatabase _db;
  final DeviceStorage _storage;
  final EventRepository _events;

  static String get _model => '${Platform.operatingSystem} ${Platform.operatingSystemVersion}'.trim();

  /// Links this device to the code's location. Order-number blocks of an earlier pairing are dropped:
  /// they belonged to that device.
  Future<void> pair(String code) async {
    final Response res;
    try {
      res = await _dio.post('api/device/pair', data: {'code': code, 'model': _model});
    } on DioException catch (e) {
      throw ApiFailure.from(e, 'Connecting', {
        400: 'Enter the 6-digit code from Bhoomy.',
        404: 'Wrong code, or older than 10 minutes. Ask for a new code in Bhoomy.',
        409: 'This location is closed. Devices can no longer be added to it.',
        429: 'Too many tries. Wait a minute, then try again.',
      });
    }
    await _db.delete(_db.numberBlocks).go();
    await _storage.save(res.data['token'], PairedDevice.fromJson(res.data));
  }

  /// Downloads what selling needs after pairing; returns counts for the pairing screen.
  Future<({int products, int promotions, int staff})> prepare() async {
    await refresh();
    final counts = await _events.localCounts(_storage.device!.location.id);
    return (products: counts.products, promotions: counts.promotions, staff: (await staff()).length);
  }

  /// Device, staff, and the snapshot when it changed. Callers at the same time share one run.
  Future<void> refresh() => _refreshing ??= _refresh().whenComplete(() => _refreshing = null);
  Future<void>? _refreshing;

  Future<void> _refresh() async {
    await refreshDevice();
    await refreshStaff();
    final eventId = _storage.device?.location.id;
    if (eventId != null) await _events.refreshSnapshot(eventId);
  }

  /// Heartbeat: device name and the location in full (the pairing reply has no location code).
  Future<void> refreshDevice() async {
    try {
      final device = (await _dio.get('api/device')).data;
      final location = (await _dio.get('api/me/event')).data;
      if (_storage.isPaired) await _storage.update(PairedDevice.fromJson({...device, 'location': location}));
    } on DioException catch (e) {
      throw ApiFailure.from(e, 'Refresh');
    }
  }

  /// Caches the location's active staff with their PIN hashes, for signing in offline.
  Future<void> refreshStaff() async {
    final List list;
    try {
      list = (await _dio.get<List>('api/device/staff')).data!;
    } on DioException catch (e) {
      throw ApiFailure.from(e, 'Loading staff');
    }
    await _db.transaction(() async {
      await _db.delete(_db.staffMembers).go();
      await _db.batch(
        (b) => b.insertAll(_db.staffMembers, [
          for (final (i, s) in list.indexed)
            StaffMembersCompanion.insert(
              id: s['id'],
              name: s['name'],
              role: s['role'],
              pinHash: s['pinHash'],
              position: i,
            ),
        ]),
      );
    });
  }

  Stream<List<StaffMember>> watchStaff() => _staffQuery().watch().map((rows) => rows.map(_toStaff).toList());

  Future<List<StaffMember>> staff() async => (await _staffQuery().get()).map(_toStaff).toList();

  SimpleSelectStatement<$StaffMembersTable, StaffRow> _staffQuery() =>
      _db.select(_db.staffMembers)..orderBy([(s) => OrderingTerm(expression: s.position)]);

  static StaffMember _toStaff(StaffRow r) =>
      StaffMember(id: r.id, name: r.name, isSupervisor: r.role == 'SUPERVISOR', pinHash: r.pinHash);
}
