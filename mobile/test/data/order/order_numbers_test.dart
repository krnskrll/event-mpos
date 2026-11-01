import 'package:dio/dio.dart';
import 'package:drift/native.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mpos/core/db/database.dart';
import 'package:mpos/data/order/order_numbers.dart';

import '../../support/fakes.dart';

void main() {
  test('tops up when low, hands out numbers in order, and stops when used up', () async {
    final db = AppDatabase(NativeDatabase.memory());
    addTearDown(db.close);
    final server = FakeServer()
      ..routes['api/me/events/EV01/number-blocks'] = (
        200,
        {'businessDate': '20261031', 'rangeStart': 41, 'rangeEnd': 42},
      );
    final numbers = OrderNumbers(db, Dio()..httpClientAdapter = server, await fakePrefs());

    await numbers.topUp('EV01');
    expect(server.requests.single.data, {'deviceId': numbers.deviceId});
    expect((await numbers.next('EV01')).number, 41);
    expect(await numbers.next('EV01'), (number: 42, businessDate: '20261031'));
    await expectLater(numbers.next('EV01'), throwsA(isA<NoOrderNumbers>()));
  });
}
