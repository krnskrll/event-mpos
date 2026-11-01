import 'dart:async';

import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/connectivity.dart';
import '../../domain/event.dart';
import '../cart/cart_notifier.dart';
import '../event/event_repository.dart';
import '../order/order_numbers.dart';
import '../order/order_uploader.dart';
import '../product/product_repository.dart';
import '../session/session.dart';

/// Background sync for an open selling session, kept alive by the app shell:
///  - every 60 s: live stock, order-number top-up, event status
///  - back online: flush the upload queue straight away
///  - event closed: drop its parked carts; queued orders keep uploading (accepted for 24 h)
final syncCoordinatorProvider = Provider<void>((ref) {
  final eventId = ref.watch(sessionEventProvider.select((e) => e?.id));
  if (eventId == null) return;

  Future<void> refresh() async {
    await ref.read(productRepositoryProvider).refreshStock(eventId);
    await ref.read(orderNumbersProvider).topUp(eventId);
    final events = await ref.read(eventRepositoryProvider).assignedEvents().catchError((_) => <AssignedEvent>[]);
    final status = events.where((e) => e.id == eventId).firstOrNull?.status;
    if (status == null || status == ref.read(sessionEventProvider)?.status) return;
    ref.read(sessionEventProvider.notifier).updateStatus(status);
    if (status == EventStatus.closed) await ref.read(cartProvider.notifier).dropEvent(eventId);
  }

  final timer = Timer.periodic(const Duration(seconds: 60), (_) => refresh());
  ref.onDispose(timer.cancel);
  ref.read(orderUploaderProvider).kick(); // orders left from an earlier run
  unawaited(refresh());

  ref.listen(onlineProvider, (previous, next) {
    if (previous?.value == false && next.value == true) {
      ref.read(orderUploaderProvider).kick();
      unawaited(refresh());
    }
  });
});
