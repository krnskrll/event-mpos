import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../support/app_harness.dart';
import '../../support/fakes.dart';

void main() {
  testWidgets('pairs with the code, downloads, then shows who can sign in', (tester) async {
    final server = FakeServer();
    serveDevice(server);
    final app = await pumpApp(tester, server, paired: false);

    expect(find.text('Enter the device code to link this device.'), findsOneWidget);
    await typeDigits(tester, '48291');
    expect(tester.widget<FilledButton>(find.widgetWithText(FilledButton, 'Connect')).onPressed, isNull);
    await typeDigits(tester, '3');
    await tester.tap(find.text('Connect'));
    await tester.pumpAndSettle();

    expect(find.text('Connected to Roadshow GI'), findsOneWidget);
    expect(find.text('Kasir 3'), findsOneWidget);
    expect(find.text('0 products, 0 promotions, 2 staff downloaded'), findsOneWidget);
    expect(app.storage.token, 't1');

    await tester.tap(find.text('Continue'));
    await tester.pumpAndSettle();
    expect(find.text("Who's selling?"), findsOneWidget);
    expect(find.text('Sari Wulandari'), findsOneWidget);
  });

  testWidgets('a wrong or expired code says what to do', (tester) async {
    final server = FakeServer()..routes['api/device/pair'] = (404, {'message': 'Pairing code not found or expired'});
    await pumpApp(tester, server, paired: false);

    await typeDigits(tester, '000000');
    await tester.tap(find.text('Connect'));
    await tester.pumpAndSettle();

    expect(find.text('Wrong code, or older than 10 minutes. Ask for a new code in Bhoomy.'), findsOneWidget);
    expect(find.text('Connect this device'), findsOneWidget);
  });

  testWidgets('a device removed in Bhoomy goes back to pairing and says why', (tester) async {
    final server = FakeServer()..validToken = 'removed';
    await pumpApp(tester, server);

    expect(find.text('Connect this device'), findsOneWidget);
    expect(find.text('This device was removed in Bhoomy. Connect it again with a new code.'), findsOneWidget);
  });
}
