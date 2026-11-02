import 'dart:io';

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:workmanager/workmanager.dart';

import 'app/app.dart';
import 'bootstrap.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  LicenseRegistry.addLicense(() async* {
    yield LicenseEntryWithLineBreaks(['Lato'], await rootBundle.loadString('assets/fonts/OFL.txt'));
  });
  if (Platform.isAndroid) await Workmanager().initialize(backgroundDispatcher);
  runApp(ProviderScope(overrides: await bootstrap(), child: const MposApp()));
}
