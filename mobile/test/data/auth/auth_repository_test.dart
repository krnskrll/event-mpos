import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mpos/core/network/api_failure.dart';
import 'package:mpos/data/auth/auth_repository.dart';

import '../../support/fakes.dart';

void main() {
  test('login stores tokens and name; failures map to BA-readable messages', () async {
    final server = FakeServer();
    final tokens = await fakeTokens();
    final repo = AuthRepository(Dio()..httpClientAdapter = server, tokens);

    await repo.login('ba@eventmpos.local', 'secret');
    expect((tokens.accessToken, tokens.userName), ('fresh', 'Bella Advisor'));

    for (final (status, message) in [(401, 'Invalid email or password'), (500, 'Login failed (500)')]) {
      server.routes['api/auth/login'] = (status, {});
      await expectLater(
        repo.login('ba@eventmpos.local', 'wrong'),
        throwsA(isA<ApiFailure>().having((e) => e.message, 'message', message)),
      );
    }
  });
}
