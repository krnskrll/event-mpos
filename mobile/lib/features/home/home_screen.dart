import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../app/status.dart';
import '../../app/theme.dart';
import '../../app/widgets.dart';
import '../../core/connectivity.dart';
import '../../core/format/money.dart';
import '../../core/notifications.dart';
import '../../core/printer/printer.dart';
import '../../core/secure/token_storage.dart';
import '../../data/auth/auth_repository.dart';
import '../../data/order/order_numbers.dart';
import '../../data/order/order_uploader.dart';
import '../../data/product/product_repository.dart';
import '../../data/promotion/promotion_repository.dart';
import '../../data/session/session.dart';
import '../../domain/event.dart';
import '../../domain/order.dart';
import '../../domain/promotion.dart';
import '../cart/draft_dialogs.dart';
import '../orders/orders_screen.dart';

final _numbersLeftProvider = FutureProvider.autoDispose((ref) {
  ref.watch(ordersProvider); // recount after each order
  return ref.watch(orderNumbersProvider).remaining(ref.watch(sessionEventProvider)!.id);
});

/// Today at a glance: start selling, today's numbers, recent orders, and (tablet) event and device.
class HomeScreen extends ConsumerStatefulWidget {
  const HomeScreen({super.key});

  @override
  ConsumerState<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends ConsumerState<HomeScreen> {
  @override
  void initState() {
    super.initState();
    requestNotificationPermission(); // queued-order updates
  }

  @override
  Widget build(BuildContext context) {
    final wide = isWide(context);
    final closed = ref.watch(sessionEventProvider.select((e) => e?.status == EventStatus.closed));
    final notice = closed ? const _ClosedCard() : (ref.isOnline ? null : const _OfflineNote());
    final main = <Widget>[
      ?notice,
      if (!closed) const _StartSelling(),
      _Stats(columns: wide ? 4 : 2),
      _RecentOrders(limit: wide ? 5 : 3),
    ];

    return Scaffold(
      body: SafeArea(
        bottom: false,
        child: wide
            ? Padding(
                padding: const EdgeInsets.fromLTRB(18, 14, 14, 14),
                child: Column(
                  children: [
                    _Header(wide: wide),
                    const SizedBox(height: 16),
                    Expanded(
                      child: Row(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Expanded(child: ListView(children: _gapped(main))),
                          const SizedBox(width: 16),
                          SizedBox(
                            width: 340,
                            child: ListView(children: _gapped(const [_EventCard(), _DeviceCard(), _PromoCard()])),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              )
            : ListView(
                padding: const EdgeInsets.fromLTRB(16, 12, 16, 16),
                children: _gapped([_Header(wide: wide), ...main]),
              ),
      ),
    );
  }
}

List<Widget> _gapped(List<Widget> children) => [
  for (final (i, child) in children.indexed) ...[if (i > 0) const SizedBox(height: 16), child],
];

String _greeting() => switch (DateTime.now().hour) {
  < 12 => 'Good morning',
  < 17 => 'Good afternoon',
  _ => 'Good evening',
};

/// "Day 3 of 14" while the event is running.
String? _eventDay(AssignedEvent event) {
  final start = DateTime.tryParse(event.startDate ?? '');
  final end = DateTime.tryParse(event.endDate ?? '');
  if (start == null || end == null) return null;
  final today = DateUtils.dateOnly(DateTime.now());
  if (today.isBefore(start) || today.isAfter(end)) return null;
  return 'Day ${today.difference(start).inDays + 1} of ${end.difference(start).inDays + 1}';
}

class _Header extends ConsumerWidget {
  const _Header({required this.wide});

  final bool wide;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final name = (ref.watch(tokenStorageProvider).userName ?? 'there').split(' ').first;
    final event = ref.watch(sessionEventProvider);
    final day = event == null ? null : _eventDay(event);
    final where = [?day, ?event?.name].join(' · ');
    return Row(
      children: [
        if (!wide) ...[const UserAvatar(), const SizedBox(width: 12)],
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                '${_greeting()}, $name',
                style: wide ? AppText.title : AppText.subhead.copyWith(fontWeight: FontWeight.w900),
              ),
              Text(
                where,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: AppText.caption.copyWith(color: Brand.slate),
              ),
            ],
          ),
        ),
        const SyncChip(),
        if (wide) ...[
          const SizedBox(width: 10),
          OutlinedButton(
            onPressed: () => ref.read(authRepositoryProvider).logout(),
            style: OutlinedButton.styleFrom(minimumSize: const Size(0, 44)),
            child: const Text('End session'),
          ),
        ],
      ],
    );
  }
}

/// Forest card that opens Sell, with Scan and Drafts shortcuts.
class _StartSelling extends ConsumerWidget {
  const _StartSelling();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final products = ref.watch(productsProvider).value?.length ?? 0;
    final drafts = ref.watch(draftsProvider).value?.length ?? 0;
    final promo = _activePromotions(ref).firstOrNull;
    final detail = ['$products products', if (promo != null) '${promo.name} is running'].join(' · ');
    Widget shortcut(IconData icon, String label, VoidCallback onTap) => Material(
      color: Colors.white.withValues(alpha: 0.12),
      borderRadius: BorderRadius.circular(12),
      child: InkWell(
        borderRadius: BorderRadius.circular(12),
        onTap: onTap,
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(icon, size: 18, color: Colors.white),
              const SizedBox(width: 8),
              Text(label, style: AppText.label.copyWith(color: Colors.white)),
            ],
          ),
        ),
      ),
    );

    return Material(
      color: Brand.forest,
      borderRadius: BorderRadius.circular(24),
      clipBehavior: Clip.antiAlias,
      child: InkWell(
        onTap: () => context.go('/sell'),
        child: Stack(
          children: [
            const Positioned(
              right: -50,
              top: -60,
              child: CircleAvatar(radius: 110, backgroundColor: Brand.forestRaised),
            ),
            Padding(
              padding: const EdgeInsets.all(20),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Expanded(
                        child: Text('Start selling', style: AppText.title.copyWith(color: Colors.white)),
                      ),
                      const CircleAvatar(
                        radius: 27,
                        backgroundColor: Brand.gold,
                        child: Icon(Icons.arrow_forward_rounded, color: Brand.onGold, size: 26),
                      ),
                    ],
                  ),
                  Text(detail, style: AppText.body.copyWith(color: Brand.greenTint)),
                  const SizedBox(height: 14),
                  Wrap(
                    spacing: 8,
                    runSpacing: 8,
                    children: [
                      shortcut(Icons.qr_code_scanner, 'Scan', () => context.push('/scan')),
                      shortcut(
                        Icons.bookmark_outline,
                        drafts > 0 ? 'Drafts · $drafts' : 'Drafts',
                        () => showDraftsDialog(context, ref),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

List<Promotion> _activePromotions(WidgetRef ref) {
  final now = DateTime.now();
  return [
    for (final p in ref.watch(promotionsProvider).value ?? const <Promotion>[])
      if ((p.startsAt == null || !now.isBefore(p.startsAt!)) && (p.endsAt == null || now.isBefore(p.endsAt!))) p,
  ];
}

/// Offline is normal: selling carries on, uploads wait.
class _OfflineNote extends StatelessWidget {
  const _OfflineNote();

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
      decoration: BoxDecoration(color: Brand.goldSoft, borderRadius: BorderRadius.circular(14)),
      child: Row(
        children: [
          const Icon(Icons.cloud_off_outlined, size: 18, color: Brand.goldInk),
          const SizedBox(width: 10),
          Expanded(
            child: Text.rich(
              TextSpan(
                children: [
                  TextSpan(
                    text: 'Keep selling. ',
                    style: AppText.label.copyWith(color: Brand.goldInk),
                  ),
                  TextSpan(
                    text: "Orders get their number and receipt now, and send when you're back online.",
                    style: AppText.caption.copyWith(color: Brand.goldInk),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}

/// The event closed: selling stopped; the job left is getting queued orders to the server.
class _ClosedCard extends ConsumerWidget {
  const _ClosedCard();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final pending = ref.watch(pendingCountProvider);
    return Container(
      padding: const EdgeInsets.all(18),
      decoration: BoxDecoration(
        color: Brand.surface,
        borderRadius: BorderRadius.circular(22),
        border: Border.all(color: Brand.goldTint),
        boxShadow: AppShadows.card,
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            pending > 0 ? 'Event closed, $pending orders to send' : 'Event closed',
            style: AppText.subhead.copyWith(fontWeight: FontWeight.w900),
          ),
          const SizedBox(height: 6),
          Text(
            pending > 0
                ? 'Selling has stopped. Keep the app open and online: orders are accepted for 24 hours after closing.'
                : 'Selling has stopped and every order has been sent. You can end the session.',
            style: AppText.body.copyWith(color: Brand.slate),
          ),
          const SizedBox(height: 14),
          Row(
            children: [
              if (pending > 0) ...[
                Expanded(
                  child: FilledButton(
                    onPressed: () => ref.read(orderUploaderProvider).kick(),
                    child: const Text('Try sending now'),
                  ),
                ),
                const SizedBox(width: 8),
              ],
              OutlinedButton(onPressed: () => context.go('/orders'), child: const Text('View orders')),
            ],
          ),
        ],
      ),
    );
  }
}

class _Stats extends ConsumerWidget {
  const _Stats({required this.columns});

  final int columns;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final orders = ref.watch(ordersProvider).value ?? const <OrderRecord>[];
    final sold = orders.where((o) => o.state == OrderState.confirmed || o.state == OrderState.pending).toList();
    final revenue = sold.fold(0.0, (sum, o) => sum + (o.total ?? 0));
    final pending = ref.watch(pendingCountProvider);
    final stats = [
      ('Revenue today', rupiah(revenue), Brand.forest),
      ('Orders', '${sold.length}', Brand.forest),
      ('Average order', sold.isEmpty ? '—' : rupiah(revenue / sold.length), Brand.forest),
      ('Still sending', '$pending', pending > 0 ? Brand.infoInk : Brand.forest),
    ];
    return LayoutBuilder(
      builder: (_, box) {
        final width = (box.maxWidth - 12 * (columns - 1)) / columns;
        return Wrap(
          spacing: 12,
          runSpacing: 12,
          children: [
            for (final (label, value, color) in stats)
              Container(
                width: width,
                padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
                decoration: BoxDecoration(
                  color: Brand.surface,
                  borderRadius: BorderRadius.circular(18),
                  border: Border.all(color: Brand.cardLine),
                  boxShadow: AppShadows.card,
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(label, style: AppText.caption),
                    const SizedBox(height: 4),
                    FittedBox(
                      fit: BoxFit.scaleDown,
                      alignment: Alignment.centerLeft,
                      child: Text(
                        value,
                        style: AppText.heading.copyWith(color: color, fontWeight: FontWeight.w900),
                      ),
                    ),
                  ],
                ),
              ),
          ],
        );
      },
    );
  }
}

class _RecentOrders extends ConsumerWidget {
  const _RecentOrders({required this.limit});

  final int limit;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final orders = (ref.watch(ordersProvider).value ?? const <OrderRecord>[]).take(limit).toList();
    return _Panel(
      title: 'Recent orders',
      action: TextButton(onPressed: () => context.go('/orders'), child: const Text('See all')),
      children: orders.isEmpty
          ? [
              Padding(
                padding: const EdgeInsets.symmetric(vertical: 16),
                child: Text('No orders yet today.', style: AppText.body.copyWith(color: Brand.muted)),
              ),
            ]
          : [
              for (final o in orders)
                InkWell(
                  onTap: () => context.push('/receipt/${o.key}'),
                  child: Padding(
                    padding: const EdgeInsets.symmetric(vertical: 10),
                    child: Row(
                      children: [
                        SizedBox(
                          width: 52,
                          child: Text(
                            o.orderNumberDisplay ?? '—',
                            style: AppText.bodyLarge.copyWith(color: Brand.forest, fontWeight: FontWeight.w900),
                          ),
                        ),
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(o.customerName, style: AppText.body.copyWith(fontWeight: FontWeight.w700)),
                              Text(
                                '${TimeOfDay.fromDateTime(o.createdAt).format(context)} · ${o.total == null ? '—' : rupiah(o.total!)}',
                                style: AppText.caption,
                              ),
                            ],
                          ),
                        ),
                        Builder(
                          builder: (_) {
                            final s = orderBadge(o.state);
                            return Pill(s.label, color: s.color, background: s.background);
                          },
                        ),
                      ],
                    ),
                  ),
                ),
            ],
    );
  }
}

/// White rounded panel with a title row and divided rows.
class _Panel extends StatelessWidget {
  const _Panel({required this.title, this.action, required this.children});

  final String title;
  final Widget? action;
  final List<Widget> children;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.fromLTRB(16, 8, 12, 8),
      decoration: BoxDecoration(
        color: Brand.surface,
        borderRadius: BorderRadius.circular(22),
        border: Border.all(color: Brand.cardLine),
        boxShadow: AppShadows.card,
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Expanded(
                child: Text(title, style: AppText.subhead.copyWith(fontWeight: FontWeight.w900)),
              ),
              ?action,
            ],
          ),
          for (final (i, child) in children.indexed) ...[if (i > 0) const Divider(), child],
        ],
      ),
    );
  }
}

/// Tablet: the event, its dates and how far along it is.
class _EventCard extends ConsumerWidget {
  const _EventCard();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final event = ref.watch(sessionEventProvider);
    if (event == null) return const SizedBox.shrink();
    final start = DateTime.tryParse(event.startDate ?? '');
    final end = DateTime.tryParse(event.endDate ?? '');
    final day = _eventDay(event);
    final progress = start == null || end == null || day == null
        ? null
        : (DateTime.now().difference(start).inDays + 1) / (end.difference(start).inDays + 1);
    final active = event.status == EventStatus.active;
    return Container(
      padding: const EdgeInsets.all(18),
      decoration: BoxDecoration(
        color: Brand.surface,
        borderRadius: BorderRadius.circular(22),
        border: Border.all(color: Brand.cardLine),
        boxShadow: AppShadows.card,
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Pill(
                active ? 'Active' : event.status.name[0].toUpperCase() + event.status.name.substring(1),
                color: active ? Brand.successInk : Brand.warningInk,
                background: active ? Brand.successSoft : Brand.warningSoft,
              ),
              const Spacer(),
              Text(event.code, style: AppText.caption),
            ],
          ),
          const SizedBox(height: 10),
          Text(event.name, style: AppText.subhead.copyWith(fontWeight: FontWeight.w900)),
          Text(
            '${event.startDate ?? '—'} → ${event.endDate ?? '—'}',
            style: AppText.caption.copyWith(color: Brand.slate),
          ),
          if (progress != null) ...[
            const SizedBox(height: 12),
            LinearProgressIndicator(
              value: progress.clamp(0, 1),
              minHeight: 8,
              borderRadius: BorderRadius.circular(999),
            ),
            const SizedBox(height: 6),
            Text(day!, style: AppText.caption),
          ],
        ],
      ),
    );
  }
}

/// Tablet: printer, order numbers left on this device, and connection.
class _DeviceCard extends ConsumerWidget {
  const _DeviceCard();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final printer = ref.watch(printerProvider).value;
    final status = ref.watch(printerStatusProvider).value;
    final left = ref.watch(_numbersLeftProvider).value;
    final online = ref.isOnline;
    Widget row(IconData icon, Color ink, Color soft, String title, String subtitle, [Widget? action]) => Padding(
      padding: const EdgeInsets.symmetric(vertical: 10),
      child: Row(
        children: [
          Container(
            width: 38,
            height: 38,
            decoration: BoxDecoration(color: soft, borderRadius: BorderRadius.circular(12)),
            child: Icon(icon, size: 19, color: ink),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  title,
                  style: AppText.body.copyWith(fontWeight: FontWeight.w700, color: Brand.forest),
                ),
                Text(subtitle, style: AppText.caption),
              ],
            ),
          ),
          ?action,
        ],
      ),
    );
    final ready = status == PrinterStatus.ready;
    return _Panel(
      title: 'This device',
      children: [
        row(
          Icons.print_outlined,
          ready ? Brand.forest : Brand.goldInk,
          ready ? Brand.greenSoft : Brand.goldSoft,
          ready ? 'Printer ready' : 'Printer: ${status?.label ?? 'checking…'}',
          printer?.name ?? '—',
          TextButton(onPressed: () => context.push('/printer'), child: const Text('Open')),
        ),
        row(Icons.tag, Brand.slate, Brand.sunken, 'Order numbers', left == null ? '—' : '$left left on this device'),
        row(
          online ? Icons.cloud_done_outlined : Icons.cloud_off_outlined,
          online ? Brand.successInk : Brand.goldInk,
          online ? Brand.successSoft : Brand.goldSoft,
          online ? 'Online' : 'Offline',
          online ? 'Stock and prices refresh every minute' : 'Selling continues; uploads wait',
        ),
      ],
    );
  }
}

/// Tablet: the promotions running now, in gold.
class _PromoCard extends ConsumerWidget {
  const _PromoCard();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final promos = _activePromotions(ref);
    if (promos.isEmpty) return const SizedBox.shrink();
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(color: Brand.goldSoft, borderRadius: BorderRadius.circular(22)),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          for (final p in promos)
            Padding(
              padding: const EdgeInsets.symmetric(vertical: 4),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Icon(Icons.sell_outlined, size: 18, color: Brand.goldInk),
                  const SizedBox(width: 10),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(p.name, style: AppText.label.copyWith(color: Brand.goldInk)),
                        Text(
                          p.minSubtotal == null
                              ? 'Applies automatically'
                              : 'Applies automatically from ${rupiah(p.minSubtotal!.toDouble())}',
                          style: AppText.caption.copyWith(color: Brand.goldInk),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
        ],
      ),
    );
  }
}
