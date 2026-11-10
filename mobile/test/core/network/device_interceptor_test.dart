import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mpos/core/network/device_interceptor.dart';
import 'package:mpos/core/secure/device_storage.dart';

import '../../support/fakes.dart';

void main() {
  late FakeServer server;
  late DeviceStorage storage;
  String? signedIn;
  late Dio dio;

  setUp(() async {
    server = FakeServer();
    storage = await fakeDevice();
    signedIn = null;
    dio = Dio()
      ..httpClientAdapter = server
      ..interceptors.add(DeviceInterceptor(storage, () => signedIn));
  });

  test('sends the device token, and X-Staff-Id only while someone is signed in', () async {
    await dio.get('api/me/event');
    expect(server.requests.last.headers['Authorization'], 'Device t1');
    expect(server.requests.last.headers.containsKey(staffHeader), isFalse);

    signedIn = budi.id;
    await dio.get('api/me/event');
    expect(server.requests.last.headers[staffHeader], budi.id);

    await dio.post('api/me/events/EV01/orders', options: Options(headers: {staffHeader: sari.id}));
    expect(server.requests.last.headers[staffHeader], sari.id);
  });

  test('a 401 means the device was removed: the pairing is dropped', () async {
    server.validToken = 'other';

    await expectLater(dio.get('api/device'), throwsA(isA<DioException>()));
    expect((storage.isPaired, storage.removed), (false, true));
  });

  test('a 401 for an older token leaves the current pairing alone', () async {
    server.routes['api/device'] = (401, {});

    await expectLater(
      dio.get('api/device', options: Options(headers: {'Authorization': 'Device old'})),
      throwsA(isA<DioException>()),
    );
    expect(storage.isPaired, isTrue);
  });
}
