import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// Opened in `main` before the app starts, then provided as an override.
final tokenStorageProvider = Provider<TokenStorage>((_) => throw UnimplementedError());

/// The signed-in account (tokens + name) in Keystore/Keychain, cached in memory. Notifies on sign-in/out.
class TokenStorage extends ChangeNotifier {
  TokenStorage._(this._store, this._values);

  static const _accessKey = 'access_token';
  static const _refreshKey = 'refresh_token';
  static const _userKey = 'user_name';

  final FlutterSecureStorage _store;
  final Map<String, String> _values;

  String? get accessToken => _values[_accessKey];
  String? get refreshToken => _values[_refreshKey];
  String? get userName => _values[_userKey];
  bool get isLoggedIn => accessToken != null && refreshToken != null;

  /// iOS keeps Keychain entries after an uninstall, so a fresh install starts signed out.
  static Future<TokenStorage> open([FlutterSecureStorage store = const FlutterSecureStorage()]) async {
    final prefs = await SharedPreferences.getInstance();
    if (prefs.getBool('installed') != true) {
      await store.deleteAll();
      await prefs.setBool('installed', true);
    }
    return TokenStorage._(store, {...await store.readAll()});
  }

  /// [userName] is set at login; a token refresh keeps the stored one.
  Future<void> save(String access, String refresh, {String? userName}) async {
    final values = {_accessKey: access, _refreshKey: refresh, _userKey: ?userName};
    for (final MapEntry(:key, :value) in values.entries) {
      await _store.write(key: key, value: value);
    }
    _values.addAll(values);
    notifyListeners();
  }

  Future<void> clear() async {
    await _store.deleteAll();
    _values.clear();
    notifyListeners();
  }
}
