import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// Opened in `main` before the app starts, then provided as an override.
final tokenStorageProvider = Provider<TokenStorage>((_) => throw UnimplementedError());

/// Access + refresh tokens in Keystore/Keychain, cached in memory. Notifies on sign-in/out.
class TokenStorage extends ChangeNotifier {
  TokenStorage._(this._store, this._access, this._refresh);

  static const _accessKey = 'access_token';
  static const _refreshKey = 'refresh_token';

  final FlutterSecureStorage _store;
  String? _access;
  String? _refresh;

  String? get accessToken => _access;
  String? get refreshToken => _refresh;
  bool get isLoggedIn => _access != null && _refresh != null;

  /// iOS keeps Keychain entries after an uninstall, so a fresh install starts signed out.
  static Future<TokenStorage> open([FlutterSecureStorage store = const FlutterSecureStorage()]) async {
    final prefs = await SharedPreferences.getInstance();
    if (prefs.getBool('installed') != true) {
      await store.deleteAll();
      await prefs.setBool('installed', true);
    }
    return TokenStorage._(store, await store.read(key: _accessKey), await store.read(key: _refreshKey));
  }

  Future<void> save(String access, String refresh) async {
    await _store.write(key: _accessKey, value: access);
    await _store.write(key: _refreshKey, value: refresh);
    _access = access;
    _refresh = refresh;
    notifyListeners();
  }

  Future<void> clear() async {
    await _store.deleteAll();
    _access = _refresh = null;
    notifyListeners();
  }
}
