import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/network/api_client.dart';
import '../../core/network/api_failure.dart';
import '../../core/secure/token_storage.dart';

final authRepositoryProvider = Provider(
  (ref) => AuthRepository(ref.watch(dioProvider), ref.watch(tokenStorageProvider)),
);

class AuthRepository {
  AuthRepository(this._dio, this._tokens);

  final Dio _dio;
  final TokenStorage _tokens;

  Future<void> login(String email, String password) async {
    try {
      final res = await _dio.post('api/auth/login', data: {'email': email, 'password': password});
      await _tokens.save(res.data['accessToken'], res.data['refreshToken'], userName: res.data['user']['fullName']);
    } on DioException catch (e) {
      throw ApiFailure.from(e, 'Login', {401: 'Invalid email or password'});
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
