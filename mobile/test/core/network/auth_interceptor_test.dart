import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mpos/core/network/auth_interceptor.dart';
import 'package:mpos/core/secure/token_storage.dart';

import '../../support/fakes.dart';

void main() {
  late FakeServer server;
  late TokenStorage tokens;
  late Dio dio;

  setUp(() async {
    server = FakeServer();
    tokens = await fakeTokens(access: 'stale', refresh: 'r1');
    final bare = Dio()..httpClientAdapter = server;
    dio = Dio()
      ..httpClientAdapter = server
      ..interceptors.add(AuthInterceptor(tokens, bare));
  });

  test('a burst of 401s refreshes once and retries every request', () async {
    final responses = await Future.wait(List.generate(3, (_) => dio.get('api/me/events')));

    expect(responses.map((r) => r.statusCode), everyElement(200));
    expect(server.refreshCalls, 1);
    expect(tokens.accessToken, 'fresh');
  });

  test('a rejected refresh signs the device out', () async {
    server.rejectRefresh = true;

    await expectLater(dio.get('api/me/events'), throwsA(isA<DioException>()));
    expect(tokens.isLoggedIn, isFalse);
  });
}
