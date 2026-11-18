import 'package:dio/dio.dart';

import '../secure/device_storage.dart';

const staffHeader = 'X-Staff-Id';
const pendingHeader = 'X-Pending-Orders';

/// Sends the device token, and X-Staff-Id for whoever is signed in unless the request names its own
/// staff (a queued order uploads as the cashier who sold it), and X-Pending-Orders: orders not
/// uploaded yet, so the back office can warn before closing an event. A 401 means Bhoomy removed
/// the device.
class DeviceInterceptor extends Interceptor {
  DeviceInterceptor(this._storage, this._staffId, {this.pending});

  final DeviceStorage _storage;
  final String? Function() _staffId;
  final Future<int> Function()? pending;

  @override
  Future<void> onRequest(RequestOptions options, RequestInterceptorHandler handler) async {
    final token = _storage.token;
    if (token != null) options.headers['Authorization'] ??= 'Device $token';
    final staff = _staffId();
    if (staff != null) options.headers[staffHeader] ??= staff;
    if (pending != null && !options.headers.containsKey(pendingHeader)) {
      try {
        options.headers[pendingHeader] = '${await pending!()}';
      } catch (_) {} // a count is a hint; never block the request
    }
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
