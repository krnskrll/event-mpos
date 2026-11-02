import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mpos/app/theme.dart';
import 'package:mpos/app/widgets.dart';

void main() {
  test('stock wording follows the level', () {
    expect(
      [stockStatus(0).label, stockStatus(5).label, stockStatus(120).label],
      ['Sold out', 'Only 5 left', '120 left'],
    );
  });

  testWidgets('stepper changes quantity and disables a side with no callback', (tester) async {
    var qty = 1;
    await tester.pumpWidget(
      MaterialApp(
        theme: appTheme,
        home: Scaffold(
          body: StatefulBuilder(
            builder: (_, setState) => QtyStepper(
              qty: qty,
              onMinus: qty > 1 ? () => setState(() => qty--) : null,
              onPlus: () => setState(() => qty++),
            ),
          ),
        ),
      ),
    );

    await tester.tap(find.byTooltip('Remove one'));
    expect(qty, 1);
    await tester.tap(find.byTooltip('Add one'));
    await tester.pump();
    expect(find.text('2'), findsOneWidget);
  });
}
