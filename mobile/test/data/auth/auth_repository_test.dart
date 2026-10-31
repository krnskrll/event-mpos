import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mpos/data/auth/auth_repository.dart';

import '../../support/fakes.dart';

void main() {
  test('login stores tokens; failures map to BA-readable messages', () async {
    final server = FakeServer();
    final tokens = await fakeTokens();
    final repo = AuthRepository(Dio()..httpClientAdapter = server, tokens);

    await repo.login('ba@eventmpos.local', 'secret');
    expect(tokens.accessToken, 'fresh');

    for (final (status, message) in [(401, 'Invalid email or password'), (500, 'Login failed (500)')]) {
      server.loginStatus = status;
      await expectLater(
        repo.login('ba@eventmpos.local', 'wrong'),
        throwsA(isA<LoginFailure>().having((e) => e.message, 'message', message)),
      );
    }
  });
}
