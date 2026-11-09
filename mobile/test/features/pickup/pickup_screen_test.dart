import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mpos/core/secure/device_storage.dart';
import 'package:mpos/data/device/device_repository.dart';
import 'package:mpos/data/pickup/pickup_repository.dart';
import 'package:mpos/domain/pickup.dart';
import 'package:mpos/features/pickup/pickup_screen.dart';

import '../../data/pickup/pickup_repository_test.dart' show pickupJson;
import '../../support/app_harness.dart';
import '../../support/fakes.dart';

class FakePickup extends PickupRepository {
  FakePickup(DeviceStorage storage) : super(Dio(), storage);

  var orders = [
    PickupOrder.fromJson(pickupJson('o1', 'NEW')),
    PickupOrder.fromJson(pickupJson('o2', 'READY', number: '#018')),
  ];
  final socket = StreamController<PickupOrder>();
  final calls = <String>[];
  int loads = 0;

  @override
  Future<List<PickupOrder>> load(String eventId) async {
    loads++;
    return orders;
  }

  @override
  Stream<PickupOrder> live(String eventId) => socket.stream;

  @override
  Future<String?> advance(String orderId, String status) async {
    calls.add('$orderId→$status');
    return null;
  }

  @override
  Future<String?> cancel(
    String orderId, {
    required String reason,
    required String supervisorId,
    required String pin,
  }) async {
    calls.add('cancel $orderId by $supervisorId');
    return pin == '482913' ? null : 'Wrong supervisor PIN';
  }
}

void main() {
  late FakePickup pickup;

  Future<void> pumpBoard(WidgetTester tester, {bool tablet = false}) async {
    tester.view
      ..physicalSize = tablet ? const Size(2048, 1536) : const Size(1170, 2532)
      ..devicePixelRatio = tablet ? 2 : 3;
    addTearDown(tester.view.reset);
    final storage = await fakeDevice();
    pickup = FakePickup(storage);
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          deviceStorageProvider.overrideWithValue(storage),
          pickupRepositoryProvider.overrideWithValue(pickup),
          staffProvider.overrideWith((_) => Stream.value([budi, sari])),
        ],
        child: const MaterialApp(home: PickupScreen()),
      ),
    );
    await tester.pumpAndSettle();
  }

  testWidgets('tablet: three columns with counts; a live update moves a card', (tester) async {
    await pumpBoard(tester, tablet: true);

    expect(find.text('New'), findsOneWidget);
    expect(find.text('Ready for pickup'), findsOneWidget);
    expect(find.text('#021'), findsOneWidget);
    expect(find.text('Live'), findsOneWidget);
    expect(find.text('Late sync'), findsNWidgets(2));

    pickup.socket.add(PickupOrder.fromJson(pickupJson('o1', 'PACKING')));
    await tester.pumpAndSettle();
    expect(find.text('Mark ready'), findsOneWidget);
    expect(find.text('Mark packing'), findsNothing);
  });

  testWidgets('phone: tabs; Mark packing moves the order on', (tester) async {
    await pumpBoard(tester);

    expect(find.text('New · 1'), findsOneWidget);
    expect(find.text('Ready for pickup · 1'), findsOneWidget);
    await tester.tap(find.text('Mark packing'));
    await tester.pumpAndSettle();

    expect(pickup.calls, ['o1→PACKING']);
    expect(find.text('New · 0'), findsOneWidget);
    expect(find.text('Packing · 1'), findsOneWidget);
  });

  testWidgets('cancel asks a supervisor; only NEW and PACKING orders can be cancelled', (tester) async {
    await pumpBoard(tester);
    await tester.tap(find.text('Ready for pickup · 1'));
    await tester.pumpAndSettle();
    expect(find.text('Cancel'), findsNothing);

    await tester.tap(find.text('New · 1'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Cancel'));
    await tester.pumpAndSettle();
    expect(find.text('Supervisor approval needed'), findsOneWidget);

    await typeDigits(tester, '1111');
    await tester.tap(find.text('Approve and cancel'));
    await tester.pumpAndSettle();
    expect(find.text('Wrong supervisor PIN'), findsOneWidget);

    await typeDigits(tester, '482913');
    await tester.tap(find.text('Approve and cancel'));
    await tester.pumpAndSettle();
    expect(pickup.calls.last, 'cancel o1 by ${sari.id}');
    expect(find.text('New · 0'), findsOneWidget);
  });

  testWidgets('when the socket fails the board polls every 10 seconds', (tester) async {
    await pumpBoard(tester);
    expect(pickup.loads, 1);

    await pickup.socket.close();
    await tester.pumpAndSettle();
    expect(find.text('Refreshing every 10 s'), findsOneWidget);

    await tester.pump(PickupBoard.pollEvery);
    await tester.pump(PickupBoard.pollEvery);
    expect(pickup.loads, 3);
  });
}
