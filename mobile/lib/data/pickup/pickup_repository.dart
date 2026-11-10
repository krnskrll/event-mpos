import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/config.dart';
import '../../core/network/api_client.dart';
import '../../core/network/api_failure.dart';
import '../../core/secure/device_storage.dart';
import '../../domain/pickup.dart';
import '../session/session.dart';

final pickupRepositoryProvider = Provider(
  (ref) => PickupRepository(ref.watch(dioProvider), ref.watch(deviceStorageProvider)),
);

class PickupRepository {
  PickupRepository(this._dio, this._storage);

  final Dio _dio;
  final DeviceStorage _storage;

  /// The orders still to hand over: new, packing and ready.
  Future<List<PickupOrder>> load(String eventId) async {
    try {
      final lists = await Future.wait([
        for (final (status, _, _) in pickupColumns)
          _dio.get<List>('api/warehouse/events/$eventId/orders', queryParameters: {'status': status}),
      ]);
      return [for (final res in lists) ...res.data!.map((o) => PickupOrder.fromJson(o))];
    } on DioException catch (e) {
      throw ApiFailure.from(e, 'Loading orders');
    }
  }

  Future<String?> advance(String orderId, String status) =>
      _send(() => _dio.patch('api/warehouse/orders/$orderId/status', data: {'status': status}));

  Future<String?> cancel(String orderId, {required String reason, required String supervisorId, required String pin}) =>
      _send(
        () => _dio.post(
          'api/warehouse/orders/$orderId/cancel',
          data: {'reason': reason, 'supervisorId': supervisorId, 'supervisorPin': pin},
        ),
      );

  /// Order changes pushed by the server as they happen.
  Stream<PickupOrder> live(String eventId) async* {
    final base = Uri.parse(apiBaseUrl);
    final url = base
        .resolve('ws/events/$eventId/orders')
        .replace(scheme: base.scheme == 'https' ? 'wss' : 'ws', queryParameters: {'token': _storage.token});
    final socket = await WebSocket.connect(url.toString());
    try {
      await for (final message in socket) {
        yield PickupOrder.fromJson(jsonDecode(message as String));
      }
    } finally {
      await socket.close();
    }
  }

  /// Returns the server's reason when it refuses, or null.
  static Future<String?> _send(Future<void> Function() call) async {
    try {
      await call();
      return null;
    } on DioException catch (e) {
      return ApiFailure.from(e, 'Update', {409: 'Someone already moved this order.'}).message;
    }
  }
}

typedef Board = ({List<PickupOrder> orders, bool live});

/// The board: loaded once, then kept current by the WebSocket, or by polling every 10 s when the
/// socket fails.
final pickupBoardProvider = AsyncNotifierProvider.autoDispose<PickupBoard, Board>(PickupBoard.new);

class PickupBoard extends AsyncNotifier<Board> {
  static const pollEvery = Duration(seconds: 10);
  Timer? _poll;

  PickupRepository get _repo => ref.read(pickupRepositoryProvider);
  String get _eventId => ref.read(sessionEventProvider)!.id;

  @override
  Future<Board> build() async {
    ref.watch(sessionEventProvider.select((e) => e?.id));
    final live = _repo.live(_eventId).listen(_apply, onError: (_) => _fallBack(), onDone: _fallBack);
    ref.onDispose(() {
      live.cancel();
      _poll?.cancel();
    });
    return (orders: await _repo.load(_eventId), live: true);
  }

  void _apply(PickupOrder order) {
    final board = state.value;
    if (board == null) return;
    final orders = [order, ...board.orders.where((o) => o.id != order.id)]
      ..sort((a, b) => b.createdAt.compareTo(a.createdAt));
    state = AsyncData((orders: orders, live: board.live));
  }

  void _fallBack() {
    _poll ??= Timer.periodic(pollEvery, (_) => refresh());
    final board = state.value;
    if (board != null) state = AsyncData((orders: board.orders, live: false));
  }

  Future<void> refresh() async {
    try {
      final orders = await _repo.load(_eventId);
      state = AsyncData((orders: orders, live: _poll == null));
    } on ApiFailure {
      return; // keep the last board; the next poll tries again
    }
  }

  /// Moves the order to its next status; returns the server's reason when it refuses.
  Future<String?> advance(PickupOrder order) async {
    final error = await _repo.advance(order.id, order.next!);
    error == null ? _apply(order.withStatus(order.next!)) : await refresh();
    return error;
  }

  Future<String?> cancel(
    PickupOrder order, {
    required String reason,
    required String supervisorId,
    required String pin,
  }) async {
    final error = await _repo.cancel(order.id, reason: reason, supervisorId: supervisorId, pin: pin);
    if (error == null) _apply(order.withStatus('CANCELLED'));
    return error;
  }
}
