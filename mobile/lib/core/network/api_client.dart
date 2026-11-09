import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../data/session/session.dart';
import '../config.dart';
import '../secure/device_storage.dart';
import 'device_interceptor.dart';

final dioProvider = Provider<Dio>((ref) {
  final options = BaseOptions(
    baseUrl: apiBaseUrl,
    connectTimeout: const Duration(seconds: 10),
    receiveTimeout: const Duration(seconds: 10),
    sendTimeout: const Duration(seconds: 10),
  );
  String? staffId() => ref.read(signedInProvider)?.id;
  return Dio(options)..interceptors.add(DeviceInterceptor(ref.watch(deviceStorageProvider), staffId));
});
