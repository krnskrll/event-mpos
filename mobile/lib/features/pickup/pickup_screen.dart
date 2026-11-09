import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../app/theme.dart';
import '../../app/widgets.dart';
import '../../core/format/money.dart';
import '../../data/pickup/pickup_repository.dart';
import '../../data/session/session.dart';
import '../../domain/pickup.dart';
import '../staff/pin_pad.dart';
import '../staff/supervisor_sheet.dart';

/// Orders to pack and hand over at this location, from every device: New, Packing, Ready for pickup.
/// Columns on tablets, tabs on phones. Any staff member can use it; cancelling needs a supervisor.
class PickupScreen extends ConsumerWidget {
  const PickupScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final board = ref.watch(pickupBoardProvider);
    final orders = board.value?.orders ?? const <PickupOrder>[];
    final columns = [
      for (final (status, title, action) in pickupColumns)
        (
          title: title,
          action: action,
          orders: [
            for (final o in orders)
              if (o.status == status) o,
          ],
        ),
    ];
    final Widget body = switch (board) {
      AsyncValue(value: null, error: final Object error) => Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          EmptyState('$error', icon: Icons.cloud_off_outlined),
          TextButton(onPressed: () => ref.invalidate(pickupBoardProvider), child: const Text('Try again')),
        ],
      ),
      AsyncValue(value: null) => const Center(child: CircularProgressIndicator()),
      _ when isWide(context) => Padding(
        padding: const EdgeInsets.all(16),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            for (final (i, c) in columns.indexed) ...[
              if (i > 0) const SizedBox(width: 12),
              Expanded(child: _Column(c.title, c.action, c.orders, index: i)),
            ],
          ],
        ),
      ),
      _ => DefaultTabController(
        length: columns.length,
        child: Column(
          children: [
            TabBar(
              labelStyle: AppText.label,
              tabs: [for (final c in columns) Tab(text: '${c.title} · ${c.orders.length}')],
            ),
            Expanded(
              child: TabBarView(
                children: [
                  for (final c in columns)
                    _Cards(c.action, c.orders, padding: const EdgeInsets.fromLTRB(16, 12, 16, 16)),
                ],
              ),
            ),
          ],
        ),
      ),
    };
    return AnnotatedRegion(
      value: SystemUiOverlayStyle.light.copyWith(statusBarColor: Colors.transparent),
      child: Scaffold(
        backgroundColor: Brand.sunken,
        body: Column(
          children: [
            _Header(live: board.value?.live),
            Expanded(child: body),
          ],
        ),
      ),
    );
  }
}

class _Header extends ConsumerWidget {
  const _Header({required this.live});

  final bool? live;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final staff = ref.watch(signedInProvider);
    final location = ref.watch(sessionEventProvider)?.name ?? '';
    return Container(
      color: Brand.forest,
      child: SafeArea(
        bottom: false,
        child: Padding(
          padding: const EdgeInsets.fromLTRB(20, 12, 12, 12),
          child: Row(
            children: [
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text('Pickup', style: AppText.title.copyWith(color: Colors.white, fontSize: 22)),
                    Row(
                      children: [
                        Flexible(
                          child: Text(
                            location,
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                            style: AppText.caption.copyWith(color: Brand.greenTint, fontSize: 13),
                          ),
                        ),
                        if (live != null) ...[
                          const SizedBox(width: 8),
                          Icon(Icons.circle, size: 8, color: live! ? Brand.greenLight : Brand.gold),
                          const SizedBox(width: 4),
                          Text(
                            live! ? 'Live' : 'Refreshing every 10 s',
                            style: AppText.caption.copyWith(color: Brand.greenTint, fontSize: 13),
                          ),
                        ],
                      ],
                    ),
                  ],
                ),
              ),
              if (staff != null && isWide(context)) ...[
                Container(
                  padding: const EdgeInsets.fromLTRB(4, 4, 12, 4),
                  decoration: BoxDecoration(color: Brand.greenPressed, borderRadius: BorderRadius.circular(999)),
                  child: Row(
                    children: [
                      StaffAvatar(staff, radius: 16),
                      const SizedBox(width: 8),
                      Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(staff.name, style: AppText.label.copyWith(color: Colors.white)),
                          Text(staff.roleLabel, style: AppText.caption.copyWith(color: Brand.greenTint)),
                        ],
                      ),
                    ],
                  ),
                ),
                const SizedBox(width: 8),
              ],
              IconButton(
                onPressed: () => ref.read(signedInProvider.notifier).lock(),
                tooltip: 'Switch cashier',
                style: IconButton.styleFrom(
                  backgroundColor: Brand.greenPressed,
                  foregroundColor: Colors.white,
                  fixedSize: const Size(44, 44),
                ),
                icon: const Icon(Icons.lock_outline),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

const _badges = [
  (Brand.goldTint, Brand.goldInk),
  (Color(0xFFCFE0F2), Color(0xFF132D4A)),
  (Color(0xFFC6EADA), Color(0xFF073827)),
];

/// Tablet column: title with its count, then the cards.
class _Column extends StatelessWidget {
  const _Column(this.title, this.action, this.orders, {required this.index});

  final String title;
  final String action;
  final List<PickupOrder> orders;
  final int index;

  @override
  Widget build(BuildContext context) {
    final (badge, ink) = _badges[index];
    return Container(
      decoration: BoxDecoration(
        color: Brand.surface,
        borderRadius: BorderRadius.circular(14),
        boxShadow: AppShadows.card,
      ),
      child: Column(
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 14, 12, 4),
            child: Row(
              children: [
                Expanded(
                  child: Text(title, style: AppText.subhead.copyWith(fontWeight: FontWeight.w900)),
                ),
                Pill('${orders.length}', color: ink, background: badge),
              ],
            ),
          ),
          Expanded(child: _Cards(action, orders, padding: const EdgeInsets.all(12))),
        ],
      ),
    );
  }
}

class _Cards extends StatelessWidget {
  const _Cards(this.action, this.orders, {required this.padding});

  final String action;
  final List<PickupOrder> orders;
  final EdgeInsets padding;

  @override
  Widget build(BuildContext context) {
    if (orders.isEmpty) return const EmptyState('No orders.', icon: Icons.inventory_2_outlined);
    return ListView.separated(
      padding: padding,
      itemCount: orders.length,
      separatorBuilder: (_, _) => const SizedBox(height: 10),
      itemBuilder: (_, i) => _OrderCard(orders[i], action: action),
    );
  }
}

class _OrderCard extends ConsumerStatefulWidget {
  const _OrderCard(this.order, {required this.action});

  final PickupOrder order;
  final String action;

  @override
  ConsumerState<_OrderCard> createState() => _OrderCardState();
}

class _OrderCardState extends ConsumerState<_OrderCard> {
  bool _busy = false;

  Future<void> _advance() async {
    setState(() => _busy = true);
    final error = await ref.read(pickupBoardProvider.notifier).advance(widget.order);
    if (!mounted) return;
    setState(() => _busy = false);
    if (error != null) ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(error)));
  }

  Future<void> _cancel() async {
    final order = widget.order;
    final board = ref.read(pickupBoardProvider.notifier);
    final messenger = ScaffoldMessenger.of(context);
    final supervisor = await showSupervisorSheet(
      context,
      subtitle: 'Cancel order ${order.number} · ${rupiah(order.total)}',
      approve: (reason, supervisor, pin) => board.cancel(order, reason: reason, supervisorId: supervisor.id, pin: pin),
    );
    if (supervisor != null) {
      messenger.showSnackBar(
        SnackBar(content: Text('Order ${order.number} cancelled, approved by ${supervisor.name}')),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final o = widget.order;
    final age = DateTime.now().difference(o.createdAt);
    final fresh = o.status == 'NEW' && age < const Duration(minutes: 2);
    final tags = [
      if (o.lateSynced) ('Late sync', Brand.sunken, Brand.ink),
      if (o.stockConflict) ('Stock conflict', Brand.errorSoft, Brand.errorInk),
    ];
    return Container(
      padding: const EdgeInsets.fromLTRB(12, 10, 12, 12),
      decoration: BoxDecoration(
        color: fresh ? Brand.goldSoft : Brand.surface,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: fresh ? Brand.gold : Brand.divider, width: fresh ? 2 : 1),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Row(
            crossAxisAlignment: CrossAxisAlignment.baseline,
            textBaseline: TextBaseline.alphabetic,
            children: [
              Expanded(child: Text(o.number, style: AppText.total.copyWith(fontSize: 24))),
              Text(
                age.inMinutes < 1 ? 'just now' : '${age.inMinutes} min',
                style: AppText.label.copyWith(color: age.inMinutes >= 15 ? Brand.errorStrong : Brand.slate),
              ),
            ],
          ),
          Text(o.customerName, style: AppText.bodyLarge),
          if (tags.isNotEmpty)
            Padding(
              padding: const EdgeInsets.only(top: 4),
              child: Wrap(
                spacing: 6,
                children: [for (final (label, bg, fg) in tags) Pill(label, color: fg, background: bg)],
              ),
            ),
          const SizedBox(height: 4),
          for (final item in o.items) Text('${item.qty} × ${item.name}', style: AppText.body),
          const SizedBox(height: 4),
          Text('Cashier: ${o.cashierName} · ${rupiah(o.total)}', style: AppText.caption),
          const SizedBox(height: 8),
          Row(
            children: [
              Expanded(
                child: FilledButton(
                  onPressed: _busy ? null : _advance,
                  style: FilledButton.styleFrom(minimumSize: const Size(0, 52)),
                  child: Text(widget.action),
                ),
              ),
              if (o.canCancel) ...[
                const SizedBox(width: 8),
                OutlinedButton(
                  onPressed: _busy ? null : _cancel,
                  style: OutlinedButton.styleFrom(
                    minimumSize: const Size(0, 52),
                    foregroundColor: Brand.errorStrong,
                    side: const BorderSide(color: Brand.errorSoft, width: 2),
                  ),
                  child: const Text('Cancel'),
                ),
              ],
            ],
          ),
        ],
      ),
    );
  }
}
