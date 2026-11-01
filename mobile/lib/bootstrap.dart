import 'package:flutter/widgets.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_riverpod/misc.dart' show Override;
import 'package:shared_preferences/shared_preferences.dart';
import 'package:workmanager/workmanager.dart';

import 'core/preferences.dart';
import 'core/secure/token_storage.dart';
import 'data/order/order_uploader.dart';

/// Platform state opened once, before any provider reads it (shared by the app and background jobs).
Future<List<Override>> bootstrap() async {
  final prefs = await SharedPreferences.getInstance();
  final tokens = await TokenStorage.open();
  return [preferencesProvider.overrideWithValue(prefs), tokenStorageProvider.overrideWithValue(tokens)];
}

/// WorkManager entry point: uploads queued orders while the app is in the background or closed.
/// Returning false asks WorkManager to retry with backoff.
@pragma('vm:entry-point')
void backgroundDispatcher() {
  Workmanager().executeTask((_, _) async {
    WidgetsFlutterBinding.ensureInitialized();
    final container = ProviderContainer(overrides: await bootstrap());
    try {
      return !await container.read(orderUploaderProvider).drainPending();
    } finally {
      container.dispose();
    }
  });
}
