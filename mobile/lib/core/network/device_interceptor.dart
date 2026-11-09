import 'package:dio/dio.dart';

import '../secure/device_storage.dart';

const staffHeader = 'X-Staff-Id';

/// Sends the device token, and X-Staff-Id for whoever is signed in unless the request names its own
/// staff (a queued order uploads as the cashier who sold it). A 401 means Bhoomy removed the device.
class DeviceInterceptor extends Interceptor {
  DeviceInterceptor(this._storage, this._staffId);

  final DeviceStorage _storage;
  final String? Function() _staffId;

  @override
  void onRequest(RequestOptions options, RequestInterceptorHandler handler) {
    final token = _storage.token;
    if (token != null) options.headers['Authorization'] ??= 'Device $token';
    final staff = _staffId();
    if (staff != null) options.headers[staffHeader] ??= staff;
    handler.next(options);
  }

  @override
  Future<void> onError(DioException err, ErrorInterceptorHandler handler) async {
    // Only the current token's rejection counts: an old request may finish after a new pairing.
    final sent = err.requestOptions.headers['Authorization'];
    if (err.response?.statusCode == 401 && _storage.token != null && sent == 'Device ${_storage.token}') {
      await _storage.clear(removed: true);
    }
    handler.next(err);
  }
}
