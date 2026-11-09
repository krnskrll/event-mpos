import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../core/connectivity.dart';
import '../core/printer/printer.dart';
import '../data/session/session.dart';
import '../domain/order.dart';
import '../features/orders/orders_screen.dart';
import 'theme.dart';
import 'widgets.dart';

/// Orders confirmed on the device that haven't reached the server yet.
final pendingCountProvider = Provider.autoDispose(
  (ref) =>
      (ref.watch(ordersProvider).value ?? const <OrderRecord>[]).where((o) => o.state == OrderState.pending).length,
);

/// The printer's current state (paper, connection), checked when a screen shows it.
final printerStatusProvider = FutureProvider.autoDispose(
  (ref) async => (await ref.watch(printerProvider.future)).status(),
);

/// Connection and upload state in one chip; replaces the old full-width offline bar.
class SyncChip extends ConsumerWidget {
  const SyncChip({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final online = ref.isOnline;
    final pending = ref.watch(pendingCountProvider);
    final (label, ink, soft, dot) = switch ((online, pending)) {
      (true, 0) => ('Synced', Brand.successInk, Brand.successSoft, Brand.success),
      (true, _) => ('Sending $pending', Brand.infoInk, Brand.infoSoft, Brand.info),
      (false, 0) => ('Offline', Brand.warningInk, Brand.warningSoft, Brand.warning),
      (false, _) => ('Offline · $pending to send', Brand.warningInk, Brand.warningSoft, Brand.warning),
    };
    return Container(
      height: 32,
      padding: const EdgeInsets.symmetric(horizontal: 11),
      decoration: BoxDecoration(color: soft, borderRadius: BorderRadius.circular(999)),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(Icons.circle, size: 8, color: dot),
          const SizedBox(width: 6),
          Text(
            label,
            style: AppText.caption.copyWith(color: ink, fontWeight: FontWeight.w700),
          ),
        ],
      ),
    );
  }
}

/// The signed-in staff member's initials in gold; opens the session menu.
class UserAvatar extends ConsumerWidget {
  const UserAvatar({super.key, this.radius = 22});

  final double radius;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final name = ref.watch(signedInProvider)?.name ?? '';
    return Semantics(
      container: true,
      button: true,
      label: 'Account and session',
      excludeSemantics: true,
      child: InkWell(
        customBorder: const CircleBorder(),
        onTap: () => showSessionMenu(context),
        child: CircleAvatar(
          radius: radius,
          backgroundColor: Brand.gold,
          child: Text(
            initials(name),
            style: AppText.label.copyWith(color: Brand.onGold, fontWeight: FontWeight.w900),
          ),
        ),
      ),
    );
  }
}

/// Who is signed in, on which device and where; printer settings; Switch cashier (locks the device).
Future<void> showSessionMenu(BuildContext context) => showModalBottomSheet(
  context: context,
  useRootNavigator: true,
  builder: (sheet) => Consumer(
    builder: (_, ref, _) {
      final staff = ref.watch(signedInProvider);
      final device = ref.watch(deviceProvider);
      return SafeArea(
        child: Padding(
          padding: const EdgeInsets.fromLTRB(12, 0, 12, 12),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              ListTile(
                leading: CircleAvatar(
                  backgroundColor: Brand.gold,
                  child: Text(
                    initials(staff?.name ?? ''),
                    style: AppText.label.copyWith(color: Brand.onGold, fontWeight: FontWeight.w900),
                  ),
                ),
                title: Text(staff?.name ?? '', style: AppText.bodyLarge.copyWith(color: Brand.forest)),
                subtitle: Text(
                  [?staff?.roleLabel, ?device?.name, ?device?.location.name].join(' · '),
                  style: AppText.caption,
                ),
              ),
              const Divider(),
              ListTile(
                leading: const Icon(Icons.print_outlined, color: Brand.green),
                title: const Text('Printer'),
                onTap: () {
                  Navigator.pop(sheet);
                  context.push('/printer');
                },
              ),
              ListTile(
                leading: const Icon(Icons.lock_outline, color: Brand.forest),
                title: const Text('Switch cashier'),
                subtitle: Text('Locks this device', style: AppText.caption),
                onTap: () {
                  Navigator.pop(sheet);
                  ref.read(signedInProvider.notifier).lock();
                },
              ),
            ],
          ),
        ),
      );
    },
  ),
);
