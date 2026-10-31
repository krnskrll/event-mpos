import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:mpos/core/secure/token_storage.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// Token storage backed by in-memory platform mocks.
Future<TokenStorage> fakeTokens({String? access, String? refresh, bool freshInstall = false}) async {
  SharedPreferences.setMockInitialValues({if (!freshInstall) 'installed': true});
  FlutterSecureStorage.setMockInitialValues({
    'access_token': ?access,
    'refresh_token': ?refresh,
  });
  return TokenStorage.open();
}

/// Answers like the backend: protected calls need [validToken]; refresh rotates it.
class FakeServer implements HttpClientAdapter {
  String validToken = 'fresh';
  bool rejectRefresh = false;
  int refreshCalls = 0;
  int? loginStatus;

  @override
  Future<ResponseBody> fetch(RequestOptions options, Stream<Uint8List>? _, Future<void>? _) async {
    switch (options.path) {
      case 'api/auth/login':
        final status = loginStatus ?? 200;
        return _json(status, status == 200 ? {'accessToken': 'fresh', 'refreshToken': 'r1'} : {});
      case 'api/auth/refresh':
        refreshCalls++;
        if (rejectRefresh) return _json(401, {});
        return _json(200, {'accessToken': validToken, 'refreshToken': 'r2'});
      default:
        final ok = options.headers['Authorization'] == 'Bearer $validToken';
        return _json(ok ? 200 : 401, {});
    }
  }

  @override
  void close({bool force = false}) {}

  static ResponseBody _json(int status, Object body) => ResponseBody.fromString(
    jsonEncode(body),
    status,
    headers: {
      Headers.contentTypeHeader: [Headers.jsonContentType],
    },
  );
}
