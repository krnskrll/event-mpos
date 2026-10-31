import 'package:dio/dio.dart';

import '../secure/token_storage.dart';

/// Adds the bearer token; on a 401 refreshes once and retries. [QueuedInterceptor] handles
/// errors one at a time, so a burst of 401s triggers a single refresh.
class AuthInterceptor extends QueuedInterceptor {
  AuthInterceptor(this._tokens, this._bare);

  final TokenStorage _tokens;

  /// No interceptors: used for refresh and retries so neither can recurse or deadlock the queue.
  final Dio _bare;

  static bool _isAuthCall(RequestOptions o) => o.path.contains('api/auth/login') || o.path.contains('api/auth/refresh');

  @override
  void onRequest(RequestOptions options, RequestInterceptorHandler handler) {
    final token = _tokens.accessToken;
    if (token != null && !_isAuthCall(options)) options.headers['Authorization'] = 'Bearer $token';
    handler.next(options);
  }

  @override
  Future<void> onError(DioException err, ErrorInterceptorHandler handler) async {
    final request = err.requestOptions;
    if (err.response?.statusCode != 401 || _isAuthCall(request)) return handler.next(err);

    // Skip the refresh when an earlier 401 in the queue already got a new token.
    final current = _tokens.accessToken;
    final alreadyRefreshed = current != null && request.headers['Authorization'] != 'Bearer $current';
    if (!alreadyRefreshed && !await _refresh()) return handler.next(err);

    request.headers['Authorization'] = 'Bearer ${_tokens.accessToken}';
    try {
      handler.resolve(await _bare.fetch(request));
    } on DioException catch (e) {
      handler.next(e);
    }
  }

  Future<bool> _refresh() async {
    final refresh = _tokens.refreshToken;
    if (refresh == null) return false;
    try {
      final res = await _bare.post('api/auth/refresh', data: {'refreshToken': refresh});
      await _tokens.save(res.data['accessToken'], res.data['refreshToken']);
      return true;
    } on DioException catch (e) {
      // The server rejected the session: sign out. A network error keeps it for the next try.
      if (e.response != null) await _tokens.clear();
      return false;
    }
  }
}
