import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// Opened in `main` (and the background isolate), then provided as an override.
final preferencesProvider = Provider<SharedPreferences>((_) => throw UnimplementedError());
