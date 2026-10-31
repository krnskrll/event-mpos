import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'app/app.dart';
import 'core/secure/token_storage.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  final tokens = await TokenStorage.open();
  runApp(ProviderScope(overrides: [tokenStorageProvider.overrideWithValue(tokens)], child: const MposApp()));
}
