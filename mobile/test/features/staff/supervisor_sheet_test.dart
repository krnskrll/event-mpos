import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mpos/core/secure/device_storage.dart';
import 'package:mpos/data/device/device_repository.dart';
import 'package:mpos/domain/staff.dart';
import 'package:mpos/features/staff/supervisor_sheet.dart';

import '../../support/app_harness.dart';
import '../../support/fakes.dart';

void main() {
  Future<List<StaffMember?>> pumpSheet(WidgetTester tester, List<StaffMember> staff, Approve approve) async {
    final results = <StaffMember?>[];
    final device = await fakeDevice();
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          deviceStorageProvider.overrideWithValue(device),
          staffProvider.overrideWith((_) => Stream.value(staff)),
        ],
        child: MaterialApp(
          home: Builder(
            builder: (context) => TextButton(
              onPressed: () async =>
                  results.add(await showSupervisorSheet(context, subtitle: '#021', approve: approve)),
              child: const Text('open'),
            ),
          ),
        ),
      ),
    );
    await tester.tap(find.text('open'));
    await tester.pumpAndSettle();
    return results;
  }

  testWidgets('no supervisor at the location: explains who can add one', (tester) async {
    await pumpSheet(tester, [budi], (_, _, _) async => null);

    expect(find.textContaining("There's no supervisor at this location yet"), findsOneWidget);
    expect(find.text('Approve and cancel'), findsNothing);
  });

  testWidgets("lists this location's supervisors; a wrong PIN shows the server's reason", (tester) async {
    tester.view
      ..physicalSize = const Size(1170, 2532)
      ..devicePixelRatio = 3;
    addTearDown(tester.view.reset);
    final sent = <(String, String, String)>[];
    final results = await pumpSheet(tester, [budi, sari], (reason, supervisor, pin) async {
      sent.add((reason, supervisor.id, pin));
      return pin == '482913' ? null : 'Wrong supervisor PIN';
    });

    expect(find.text('Supervisors at Roadshow GI'), findsOneWidget);
    expect(find.text('Sari Wulandari'), findsOneWidget);
    expect(find.text('Budi Santoso'), findsNothing);

    await typeDigits(tester, '1111');
    await tester.tap(find.text('Approve and cancel'));
    await tester.pumpAndSettle();
    expect(find.text('Wrong supervisor PIN'), findsOneWidget);

    await typeDigits(tester, '482913');
    await tester.tap(find.text('Approve and cancel'));
    await tester.pumpAndSettle();
    expect(sent.last, ('Customer cancelled', sari.id, '482913'));
    expect(results.single?.id, sari.id);
    expect(find.text('Supervisor approval needed'), findsNothing);
  });
}
