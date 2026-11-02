import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../data/cart/cart_notifier.dart';
import '../data/order/order_uploader.dart';
import '../data/session/session.dart';
import '../data/sync/sync_coordinator.dart';
import '../domain/event.dart';
import '../domain/order.dart';
import '../features/orders/orders_screen.dart';
import 'logo.dart';
import 'theme.dart';

/// The selling session's frame: bottom bar on phones, side rail on tablets. Keeps background sync
/// running and flushes the upload queue whenever the app comes back to the foreground.
class AppShell extends ConsumerStatefulWidget {
  const AppShell(this.shell, {super.key});

  final StatefulNavigationShell shell;

  @override
  ConsumerState<AppShell> createState() => _AppShellState();
}

class _AppShellState extends ConsumerState<AppShell> {
  late final _lifecycle = AppLifecycleListener(onResume: () => ref.read(orderUploaderProvider).kick());

  @override
  void dispose() {
    _lifecycle.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    ref.watch(syncCoordinatorProvider);
    final cartCount = ref.watch(cartProvider.select((c) => c.count));
    final destinations = [
      (Icons.home_outlined, 'Home', null),
      (Icons.shopping_bag_outlined, 'Catalogue', null),
      (Icons.shopping_cart_outlined, 'Cart', cartCount),
      (Icons.assignment_outlined, 'Orders', null),
    ];
    Widget icon(IconData data, int? badge) =>
        Badge(isLabelVisible: (badge ?? 0) > 0, label: Text('$badge'), child: Icon(data));
    void select(int i) => widget.shell.goBranch(i, initialLocation: i == widget.shell.currentIndex);
    final body = Column(
      children: [
        const _ClosedEventBanner(),
        Expanded(child: widget.shell),
      ],
    );

    if (MediaQuery.sizeOf(context).width >= 840) {
      return Scaffold(
        body: Row(
          children: [
            NavigationRail(
              selectedIndex: widget.shell.currentIndex,
              onDestinationSelected: select,
              labelType: NavigationRailLabelType.all,
              leading: const Padding(
                padding: EdgeInsets.fromLTRB(0, 8, 0, 24),
                child: BhoomyLogo(size: 44, tile: false),
              ),
              destinations: [
                for (final (data, label, badge) in destinations)
                  NavigationRailDestination(icon: icon(data, badge), label: Text(label)),
              ],
            ),
            Expanded(child: body),
          ],
        ),
      );
    }
    return Scaffold(
      body: body,
      bottomNavigationBar: NavigationBar(
        selectedIndex: widget.shell.currentIndex,
        onDestinationSelected: select,
        destinations: [
          for (final (data, label, badge) in destinations) NavigationDestination(icon: icon(data, badge), label: label),
        ],
      ),
    );
  }
}

/// The event closed while orders are still waiting to upload: keep the app open and online.
class _ClosedEventBanner extends ConsumerWidget {
  const _ClosedEventBanner();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final closed = ref.watch(sessionEventProvider)?.status == EventStatus.closed;
    final pending = (ref.watch(ordersProvider).value ?? const []).where((o) => o.state == OrderState.pending).length;
    if (!closed || pending == 0) return const SizedBox.shrink();
    return Material(
      color: Brand.forest,
      child: SafeArea(
        bottom: false,
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
          child: Text(
            "Event closed — $pending order(s) still syncing. Keep the app open and online; they're accepted for 24h.",
            style: AppText.caption.copyWith(color: Colors.white, fontWeight: FontWeight.w700),
          ),
        ),
      ),
    );
  }
}
