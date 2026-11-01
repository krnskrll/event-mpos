import 'package:dio/dio.dart';

/// A failed API call with a message the BA can read.
class ApiFailure implements Exception {
  const ApiFailure(this.message);
  final String message;

  /// [messages] overrides the text per HTTP status; otherwise the server's own message is used.
  factory ApiFailure.from(DioException e, String action, [Map<int, String> messages = const {}]) {
    final status = e.response?.statusCode;
    if (status == null) return const ApiFailure("Can't reach the server. Check the venue network.");
    final data = e.response?.data;
    final serverMessage = data is Map ? data['message'] as String? : null;
    return ApiFailure(messages[status] ?? serverMessage ?? '$action failed ($status)');
  }

  @override
  String toString() => message;
}
