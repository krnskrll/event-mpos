import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mpos/data/session/session.dart';

import '../../support/app_harness.dart';
import '../../support/fakes.dart';

void main() {
  late FakeServer server;

  setUp(() {
    server = FakeServer();
    serveDevice(server);
  });

  Future<void> enterPin(WidgetTester tester, String pin) async {
    await typeDigits(tester, pin);
    await tester.tap(find.text('Sign in'));
    await tester.pumpAndSettle();
  }

  testWidgets('a paired device opens locked; the right PIN signs in and X-Staff-Id follows', (tester) async {
    final app = await pumpApp(tester, server);

    expect(find.text("Who's selling?"), findsOneWidget);
    expect(find.text('Roadshow GI'), findsOneWidget);
    expect(find.text('Device: Kasir 3', findRichText: true), findsOneWidget); // renamed by the heartbeat
    await tester.tap(find.text('Budi Santoso'));
    await tester.pumpAndSettle();
    expect(find.text('Enter your PIN'), findsOneWidget);

    await enterPin(tester, '9999');
    expect(find.text('Wrong PIN. 4 tries left.'), findsOneWidget);
    expect(app.container.read(signedInProvider), isNull);

    await enterPin(tester, '1234');
    expect(app.container.read(signedInProvider)?.id, budi.id);
    expect(find.textContaining('Budi'), findsWidgets);
    expect(find.text("Who's selling?"), findsNothing);
    expect(server.requests.last.headers['X-Staff-Id'], budi.id);
  });

  testWidgets('5 wrong PINs make that person wait; others can still sign in', (tester) async {
    await pumpApp(tester, server);
    await tester.tap(find.text('Budi Santoso'));
    await tester.pumpAndSettle();

    for (var i = 0; i < 5; i++) {
      await enterPin(tester, '0000');
    }
    expect(find.textContaining('Too many wrong tries. Try again in'), findsOneWidget);
    expect(tester.widget<FilledButton>(find.widgetWithText(FilledButton, 'Sign in')).onPressed, isNull);

    await tester.tap(find.byTooltip('Back to staff list'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Sari Wulandari'));
    await tester.pumpAndSettle();
    expect(find.textContaining('Too many wrong tries'), findsNothing);
  });

  testWidgets('5 idle minutes lock the device', (tester) async {
    final app = await pumpApp(tester, server);
    await tester.tap(find.text('Sari Wulandari'));
    await tester.pumpAndSettle();
    await enterPin(tester, '482913');
    expect(app.container.read(signedInProvider)?.id, sari.id);

    await tester.pump(idleLock);
    await tester.pumpAndSettle();
    expect(app.container.read(signedInProvider), isNull);
    expect(find.text("Who's selling?"), findsOneWidget);
  });
}
