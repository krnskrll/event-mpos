import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../core/printer/printer.dart';
import '../data/order/order_uploader.dart';
import '../data/session/session.dart';
import '../data/sync/sync_coordinator.dart';
import '../domain/event.dart';
import 'logo.dart';
import 'status.dart';
import 'theme.dart';
import 'widgets.dart';

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
    final pending = ref.watch(pendingCountProvider);
    final closed = ref.watch(sessionEventProvider.select((e) => e?.status == EventStatus.closed));
    // A closed event stops selling: leave the Sell tab if it's open.
    ref.listen(sessionEventProvider.select((e) => e?.status), (_, status) {
      if (status == EventStatus.closed && widget.shell.currentIndex == 1) widget.shell.goBranch(0);
    });
    final destinations = [
      (Icons.home_outlined, 'Home', 0, true),
      (Icons.grid_view_rounded, 'Sell', 0, !closed),
      (Icons.receipt_long_outlined, 'Orders', pending, true),
      (Icons.inventory_2_outlined, 'Pickup', 0, true),
    ];
    Widget icon(IconData data, int badge) => Badge(
      isLabelVisible: badge > 0,
      label: Text('$badge'),
      backgroundColor: Brand.gold,
      textColor: Brand.onGold,
      child: Icon(data),
    );
    void select(int i) => widget.shell.goBranch(i, initialLocation: i == widget.shell.currentIndex);

    if (isWide(context)) {
      return Scaffold(
        body: SafeArea(
          child: Row(
            children: [
              Container(
                margin: const EdgeInsets.fromLTRB(14, 14, 0, 14),
                clipBehavior: Clip.antiAlias,
                decoration: BoxDecoration(color: Brand.forest, borderRadius: BorderRadius.circular(24)),
                child: NavigationRail(
                  backgroundColor: Colors.transparent,
                  selectedIndex: widget.shell.currentIndex,
                  onDestinationSelected: select,
                  labelType: NavigationRailLabelType.all,
                  leading: const Padding(padding: EdgeInsets.fromLTRB(0, 10, 0, 24), child: BhoomyLogo(size: 46)),
                  trailing: const Expanded(
                    child: Align(
                      alignment: Alignment.bottomCenter,
                      child: Padding(
                        padding: EdgeInsets.only(bottom: 16),
                        child: Column(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            _PrinterButton(),
                            SizedBox(height: 10),
                            _LockButton(),
                            SizedBox(height: 10),
                            UserAvatar(),
                          ],
                        ),
                      ),
                    ),
                  ),
                  destinations: [
                    for (final (data, label, badge, enabled) in destinations)
                      NavigationRailDestination(icon: icon(data, badge), label: Text(label), disabled: !enabled),
                  ],
                ),
              ),
              Expanded(child: widget.shell),
            ],
          ),
        ),
      );
    }
    return Scaffold(
      body: widget.shell,
      bottomNavigationBar: SafeArea(
        top: false,
        child: Container(
          margin: const EdgeInsets.fromLTRB(14, 0, 14, 10),
          clipBehavior: Clip.antiAlias,
          decoration: BoxDecoration(
            color: Brand.surface,
            borderRadius: BorderRadius.circular(22),
            border: Border.all(color: Brand.cardLine),
            boxShadow: AppShadows.panel,
          ),
          child: NavigationBar(
            height: 66,
            backgroundColor: Colors.transparent,
            selectedIndex: widget.shell.currentIndex,
            onDestinationSelected: select,
            destinations: [
              for (final (data, label, badge, enabled) in destinations)
                NavigationDestination(icon: icon(data, badge), label: label, enabled: enabled),
            ],
          ),
        ),
      ),
    );
  }
}

/// Switch cashier: locks the device back to the staff list.
class _LockButton extends ConsumerWidget {
  const _LockButton();

  @override
  Widget build(BuildContext context, WidgetRef ref) => IconButton(
    onPressed: () => ref.read(signedInProvider.notifier).lock(),
    tooltip: 'Switch cashier',
    style: IconButton.styleFrom(
      backgroundColor: Brand.forestRaised,
      foregroundColor: Brand.greenTint,
      fixedSize: const Size(48, 48),
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
    ),
    icon: const Icon(Icons.lock_outline),
  );
}

/// Printer shortcut with a status dot: green when ready, gold otherwise.
class _PrinterButton extends ConsumerWidget {
  const _PrinterButton();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final ready = ref.watch(printerStatusProvider).value == PrinterStatus.ready;
    return IconButton(
      onPressed: () => context.push('/printer'),
      tooltip: ready ? 'Printer ready' : 'Printer needs attention',
      style: IconButton.styleFrom(
        backgroundColor: Brand.forestRaised,
        foregroundColor: Brand.greenTint,
        fixedSize: const Size(48, 48),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
      ),
      icon: Badge(
        smallSize: 8,
        backgroundColor: ready ? Brand.greenLight : Brand.gold,
        child: const Icon(Icons.print_outlined),
      ),
    );
  }
}
