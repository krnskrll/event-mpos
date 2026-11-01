import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mpos/app/app.dart';
import 'package:mpos/core/secure/token_storage.dart';

import 'support/fakes.dart';

void main() {
  testWidgets('a signed-out user starts on the login screen', (tester) async {
    final tokens = await fakeTokens();
    await tester.pumpWidget(
      ProviderScope(overrides: [tokenStorageProvider.overrideWithValue(tokens)], child: const MposApp()),
    );
    await tester.pumpAndSettle();

    expect(find.text('Welcome Back'), findsOneWidget);
  });
}
