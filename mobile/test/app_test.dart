import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mpos/app/app.dart';
import 'package:mpos/core/secure/device_storage.dart';

import 'support/fakes.dart';

void main() {
  testWidgets('a device that is not paired starts on the pairing screen', (tester) async {
    final device = await fakeDevice(paired: false);
    await tester.pumpWidget(
      ProviderScope(overrides: [deviceStorageProvider.overrideWithValue(device)], child: const MposApp()),
    );
    await tester.pumpAndSettle();

    expect(find.text('Connect this device'), findsOneWidget);
  });
}
