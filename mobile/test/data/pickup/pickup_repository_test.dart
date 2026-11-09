import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mpos/core/network/device_interceptor.dart';
import 'package:mpos/data/pickup/pickup_repository.dart';

import '../../support/fakes.dart';

Map<String, Object> pickupJson(String id, String status, {String number = '#021'}) => {
  'id': id,
  'orderNumber': 21,
  'orderNumberDisplay': number,
  'referenceId': 'ORD-$id',
  'customerName': 'Aisha Rahman',
  'baName': 'Budi Santoso',
  'items': [
    {'name': 'Velvet Lipstick', 'sku': 'VL-1', 'qty': 2},
  ],
  'status': status,
  'total': 287000,
  'createdAt': DateTime.now().toUtc().toIso8601String(),
  'stockConflict': false,
  'lateSynced': true,
};

void main() {
  test('loads the open columns, moves and cancels orders as the signed-in staff member', () async {
    final server = FakeServer()..routes['api/warehouse/events/EV01/orders'] = (200, [pickupJson('o1', 'NEW')]);
    final storage = await fakeDevice();
    final repo = PickupRepository(
      Dio()
        ..httpClientAdapter = server
        ..interceptors.add(DeviceInterceptor(storage, () => budi.id)),
      storage,
    );

    final orders = await repo.load('EV01');
    expect(orders, hasLength(3)); // the fake answers each status query with the same order
    expect(server.requests.map((r) => r.queryParameters['status']), ['NEW', 'PACKING', 'READY']);
    expect((orders.first.number, orders.first.items.single.qty, orders.first.lateSynced), ('#021', 2, true));

    expect(await repo.advance('o1', 'PACKING'), isNull);
    expect(server.requests.last.method, 'PATCH');
    expect(server.requests.last.data, {'status': 'PACKING'});
    expect(server.requests.last.headers[staffHeader], budi.id);

    server.routes['api/warehouse/orders/o1/cancel'] = (403, {'message': 'Wrong supervisor PIN'});
    expect(await repo.cancel('o1', reason: 'Out of stock', supervisorId: sari.id, pin: '0000'), 'Wrong supervisor PIN');
    expect(server.requests.last.data, {'reason': 'Out of stock', 'supervisorId': sari.id, 'supervisorPin': '0000'});
  });
}
