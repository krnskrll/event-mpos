import 'dart:convert';

import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../../domain/device.dart';

/// Opened in `main` before the app starts, then provided as an override.
final deviceStorageProvider = Provider<DeviceStorage>((_) => throw UnimplementedError());

/// The pairing (device token, device, store and location) in Keystore/Keychain, cached in memory.
/// Notifies when the device is paired, updated or removed.
class DeviceStorage extends ChangeNotifier {
  DeviceStorage._(this._store, this._token, this._device);

  static const _tokenKey = 'device_token';
  static const _deviceKey = 'device';

  final FlutterSecureStorage _store;
  String? _token;
  PairedDevice? _device;

  /// True after the server stopped accepting this device, so pairing can say why it is back.
  bool removed = false;

  String? get token => _token;
  PairedDevice? get device => _device;
  bool get isPaired => _token != null && _device != null;

  /// iOS keeps Keychain entries after an uninstall, so a fresh install starts unpaired.
  static Future<DeviceStorage> open([FlutterSecureStorage store = const FlutterSecureStorage()]) async {
    final prefs = await SharedPreferences.getInstance();
    if (prefs.getBool('installed') != true) {
      await store.deleteAll();
      await prefs.setBool('installed', true);
    }
    final values = await store.readAll();
    final device = values[_deviceKey];
    return DeviceStorage._(store, values[_tokenKey], device == null ? null : PairedDevice.fromJson(jsonDecode(device)));
  }

  Future<void> save(String token, PairedDevice device) async {
    await _store.write(key: _tokenKey, value: token);
    _token = token;
    removed = false;
    await update(device);
  }

  Future<void> update(PairedDevice device) async {
    await _store.write(key: _deviceKey, value: jsonEncode(device.toJson()));
    _device = device;
    notifyListeners();
  }

  Future<void> clear({bool removed = false}) async {
    await _store.deleteAll();
    _token = null;
    _device = null;
    this.removed = removed;
    notifyListeners();
  }
}
