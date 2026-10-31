import 'package:connectivity_plus/connectivity_plus.dart';
import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'config.dart';

/// True while the backend is reachable. Every network change triggers a health check, so venue
/// Wi-Fi without internet reads as offline. Until the first check completes, assume online.
final onlineProvider = StreamProvider<bool>((ref) async* {
  final connectivity = Connectivity();
  final dio = Dio(BaseOptions(baseUrl: apiBaseUrl, connectTimeout: const Duration(seconds: 3)));

  Future<bool> reachable(List<ConnectivityResult> network) async {
    if (network.every((r) => r == ConnectivityResult.none)) return false;
    try {
      await dio.get('api/health');
      return true;
    } on DioException {
      return false;
    }
  }

  yield await reachable(await connectivity.checkConnectivity());
  await for (final network in connectivity.onConnectivityChanged) {
    yield await reachable(network);
  }
});

extension OnlineRef on WidgetRef {
  bool get isOnline => watch(onlineProvider).value ?? true;
}
