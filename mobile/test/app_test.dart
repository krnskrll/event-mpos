import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mpos/app/app.dart';

void main() {
  testWidgets('app starts', (tester) async {
    await tester.pumpWidget(const ProviderScope(child: MposApp()));
    expect(find.text('event-mpos'), findsOneWidget);
  });
}
