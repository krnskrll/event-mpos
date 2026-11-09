import 'dart:async';

import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/connectivity.dart';
import '../../domain/event.dart';
import '../cart/cart_notifier.dart';
import '../order/order_numbers.dart';
import '../order/order_uploader.dart';
import '../product/product_repository.dart';
import '../session/session.dart';

/// Background sync while someone is selling, kept alive by the app shell (the device heartbeat that
/// tracks the location's status runs app-wide, see `deviceSyncProvider`):
///  - every 60 s: live stock, order-number top-up
///  - back online: flush the upload queue straight away
///  - event closed: drop its parked carts; queued orders keep uploading (accepted for 24 h)
final syncCoordinatorProvider = Provider<void>((ref) {
  final eventId = ref.watch(sessionEventProvider.select((e) => e?.id));
  if (eventId == null) return;

  Future<void> refresh() async {
    await ref.read(productRepositoryProvider).refreshStock(eventId);
    await ref.read(orderNumbersProvider).topUp(eventId);
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
  ref.listen(sessionEventProvider.select((e) => e?.status), (previous, status) {
    if (status == EventStatus.closed && previous != status) ref.read(cartProvider.notifier).dropEvent(eventId);
  }, fireImmediately: true);
});
