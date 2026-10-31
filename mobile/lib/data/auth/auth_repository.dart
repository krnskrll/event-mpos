import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/network/api_client.dart';
import '../../core/secure/token_storage.dart';

final authRepositoryProvider = Provider(
  (ref) => AuthRepository(ref.watch(dioProvider), ref.watch(tokenStorageProvider)),
);

class LoginFailure implements Exception {
  const LoginFailure(this.message);
  final String message;
}

class AuthRepository {
  AuthRepository(this._dio, this._tokens);

  final Dio _dio;
  final TokenStorage _tokens;

  Future<void> login(String email, String password) async {
    try {
      final res = await _dio.post('api/auth/login', data: {'email': email, 'password': password});
      await _tokens.save(res.data['accessToken'], res.data['refreshToken']);
    } on DioException catch (e) {
      throw LoginFailure(switch (e.response?.statusCode) {
        null => "Can't reach the server. Check the venue network.",
        401 => 'Invalid email or password',
        final code => 'Login failed ($code)',
      });
    }
  }

  /// Server-side revocation is best effort; the device always signs out.
  Future<void> logout() async {
    try {
      await _dio.post('api/auth/logout');
    } on DioException catch (_) {}
    await _tokens.clear();
  }
}
