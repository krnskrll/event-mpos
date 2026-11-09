import 'package:dio/dio.dart';
import 'package:drift/drift.dart' show DatabaseConnection, driftRuntimeOptions;
import 'package:drift/native.dart';
import 'package:flutter/services.dart';
import 'package:flutter_local_notifications/flutter_local_notifications.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_riverpod/misc.dart' show Override;
import 'package:flutter_test/flutter_test.dart';
import 'package:mpos/app/app.dart';
import 'package:mpos/core/connectivity.dart';
import 'package:mpos/core/db/database.dart';
import 'package:mpos/core/network/api_client.dart';
import 'package:mpos/core/network/device_interceptor.dart';
import 'package:mpos/core/preferences.dart';
import 'package:mpos/core/secure/device_storage.dart';
import 'package:mpos/data/session/session.dart';
import 'package:mpos/features/staff/pin_pad.dart';

import 'fakes.dart';

/// The whole app against [server], an in-memory database and mocked platform storage.
/// PINs are compared as plain text: [pinOf] maps a hash to its PIN.
Future<({AppDatabase db, DeviceStorage storage, ProviderContainer container})> pumpApp(
  WidgetTester tester,
  FakeServer server, {
  bool paired = true,
  List<Override> overrides = const [],
}) async {
  tester.view
    ..physicalSize =
        const Size(1170, 2532) // a 390 × 844 phone
    ..devicePixelRatio = 3;
  addTearDown(tester.view.reset);
  AndroidFlutterLocalNotificationsPlugin.registerWith(); // Home asks for notification permission
  tester.binding.defaultBinaryMessenger.setMockMethodCallHandler(
    const MethodChannel('dexterous.com/flutter/local_notifications'),
    (_) async => true,
  );
  driftRuntimeOptions.dontWarnAboutMultipleDatabases = true; // one in-memory database per test
  final storage = await fakeDevice(paired: paired);
  final prefs = await fakePrefs();
  final db = AppDatabase(DatabaseConnection(NativeDatabase.memory(), closeStreamsSynchronously: true));
  await tester.pumpWidget(
    ProviderScope(
      overrides: [
        deviceStorageProvider.overrideWithValue(storage),
        preferencesProvider.overrideWithValue(prefs),
        databaseProvider.overrideWithValue(db),
        onlineProvider.overrideWith((_) => Stream.value(true)),
        pinCheckProvider.overrideWithValue((pin, hash) async => pinOf[hash] == pin),
        dioProvider.overrideWith(
          (ref) => Dio()
            ..httpClientAdapter = server
            ..interceptors.add(DeviceInterceptor(storage, () => ref.read(signedInProvider)?.id)),
        ),
        ...overrides,
      ],
      child: const MposApp(),
    ),
  );
  await tester.pumpAndSettle();
  return (db: db, storage: storage, container: ProviderScope.containerOf(tester.element(find.byType(MposApp))));
}

const pinOf = {pin1234: '1234', pin482913: '482913'};

/// Taps digits on the visible number pad.
Future<void> typeDigits(WidgetTester tester, String digits) async {
  for (final d in digits.split('')) {
    await tester.tap(find.descendant(of: find.byType(NumberPad), matching: find.text(d)).last);
    await tester.pump();
  }
}
