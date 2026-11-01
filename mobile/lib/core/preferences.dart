import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:uuid/uuid.dart';

/// Opened in `main` (and the background isolate), then provided as an override.
final preferencesProvider = Provider<SharedPreferences>((_) => throw UnimplementedError());

/// A random id per install. Kept in preferences, not the Keychain, so a reinstall is a new device
/// and never reuses another install's order numbers.
String deviceIdOf(SharedPreferences prefs) {
  final id = prefs.getString('device_id') ?? const Uuid().v4();
  prefs.setString('device_id', id);
  return id;
}
