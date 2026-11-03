import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../app/status.dart';
import '../../app/theme.dart';
import '../../app/widgets.dart';
import '../../core/connectivity.dart';
import '../../core/format/money.dart';
import '../../data/order/order_repository.dart';
import '../../data/session/session.dart';
import '../../domain/order.dart';
import 'order_detail.dart';

final ordersProvider = StreamProvider.autoDispose(
  (ref) => ref.watch(orderRepositoryProvider).watchOrders(ref.watch(sessionEventProvider)!.id),
);

const _filters = <(String, OrderState?)>[
  ('All', null),
  ('Confirmed', OrderState.confirmed),
  ('Sending', OrderState.pending),
  ('Failed', OrderState.failed),
  ('Cancelled', OrderState.cancelled),
];

/// Status pill text and colours, shared by Orders and Home.
({String label, Color color, Color background}) orderBadge(OrderState state) => switch (state) {
  OrderState.pending => (label: 'Sending', color: Brand.infoInk, background: Brand.infoSoft),
  OrderState.confirmed => (label: 'Confirmed', color: Brand.successInk, background: Brand.successSoft),
  OrderState.failed => (label: 'Failed', color: Brand.errorInk, background: Brand.errorSoft),
  OrderState.cancelled => (label: 'Cancelled', color: Brand.slate, background: Brand.sunken),
};

/// Today's orders. Tablets show the list beside the selected order; phones open an order full screen.
class OrdersScreen extends ConsumerStatefulWidget {
  const OrdersScreen({super.key});

  @override
  ConsumerState<OrdersScreen> createState() => _OrdersScreenState();
}

class _OrdersScreenState extends ConsumerState<OrdersScreen> {
  OrderState? _filter;
  String _query = '';
  String? _selected;

  bool _matches(OrderRecord o) {
    final q = _query.toLowerCase();
    return (_filter == null || o.state == _filter) &&
        (q.isEmpty || (o.orderNumberDisplay ?? '').contains(q) || o.customerName.toLowerCase().contains(q));
  }

  @override
  Widget build(BuildContext context) {
    final wide = isWide(context);
    final orders = ref.watch(ordersProvider).value ?? const <OrderRecord>[];
    final shown = orders.where(_matches).toList();
    final sold = orders.where((o) => o.state == OrderState.confirmed || o.state == OrderState.pending);
    final revenue = sold.fold(0.0, (sum, o) => sum + (o.total ?? 0));
    final selected = shown.where((o) => o.key == _selected).firstOrNull ?? shown.firstOrNull;

    final list = Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Row(
          children: [
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text("Today's orders", style: AppText.title),
                  Text(
                    '${sold.length} orders · ${rupiah(revenue)}',
                    style: AppText.caption.copyWith(color: Brand.slate),
                  ),
                ],
              ),
            ),
            const SyncChip(),
          ],
        ),
        const SizedBox(height: 12),
        TextField(
          onChanged: (v) => setState(() => _query = v.trim()),
          decoration: const InputDecoration(
            hintText: 'Search order number or customer',
            prefixIcon: Icon(Icons.search, color: Brand.muted),
          ),
        ),
        SizedBox(
          height: 58,
          child: ListView(
            scrollDirection: Axis.horizontal,
            padding: const EdgeInsets.symmetric(vertical: 8),
            children: [
              for (final (label, state) in _filters)
                Padding(
                  padding: const EdgeInsets.only(right: 8),
                  child: CountChip(
                    label: label,
                    count: orders.where((o) => state == null || o.state == state).length,
                    selected: _filter == state,
                    onTap: () => setState(() => _filter = state),
                  ),
                ),
            ],
          ),
        ),
        Expanded(
          child: shown.isEmpty
              ? EmptyState(
                  orders.isEmpty ? 'No orders yet. Start selling!' : 'No orders match.',
                  icon: Icons.receipt_long_outlined,
                )
              : ListView.separated(
                  padding: const EdgeInsets.only(bottom: 16),
                  itemCount: shown.length,
                  separatorBuilder: (_, _) => const SizedBox(height: 8),
                  itemBuilder: (_, i) {
                    final o = shown[i];
                    return _OrderRow(
                      o,
                      selected: wide && o.key == selected?.key,
                      onTap: () => wide ? setState(() => _selected = o.key) : context.push('/order/${o.key}'),
                    );
                  },
                ),
        ),
      ],
    );

    return Scaffold(
      body: SafeArea(
        bottom: false,
        child: Padding(
          padding: wide ? const EdgeInsets.fromLTRB(18, 14, 14, 14) : const EdgeInsets.fromLTRB(16, 12, 16, 0),
          child: wide
              ? Row(
                  children: [
                    Expanded(child: list),
                    const SizedBox(width: 16),
                    Container(
                      width: 400,
                      padding: const EdgeInsets.all(20),
                      decoration: BoxDecoration(
                        color: Brand.surface,
                        borderRadius: BorderRadius.circular(24),
                        border: Border.all(color: Brand.divider),
                        boxShadow: AppShadows.panel,
                      ),
                      child: selected == null
                          ? const EmptyState('Select an order to see it here.', icon: Icons.touch_app_outlined)
                          : OrderDetail(selected),
                    ),
                  ],
                )
              : list,
        ),
      ),
    );
  }
}

/// Number, customer, why it's in its state, status and total.
class _OrderRow extends ConsumerWidget {
  const _OrderRow(this.order, {required this.selected, required this.onTap});

  final OrderRecord order;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final badge = orderBadge(order.state);
    final time = TimeOfDay.fromDateTime(order.createdAt).format(context);
    final items = itemCount(order.itemCount);
    final (meta, metaColor) = switch (order.state) {
      OrderState.pending => ('$time · ${ref.isOnline ? 'sending now' : 'waiting for internet'}', Brand.muted),
      OrderState.failed => ('Not accepted: ${order.error ?? 'rejected'}', Brand.errorInk),
      OrderState.cancelled => ('$time · cancelled, stock returned', Brand.muted),
      OrderState.confirmed => ('$time · $items', Brand.muted),
    };
    return Material(
      color: selected ? Brand.greenSoft : Brand.surface,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(18),
        side: BorderSide(color: selected ? Brand.greenTint : Brand.cardLine),
      ),
      child: InkWell(
        borderRadius: BorderRadius.circular(18),
        onTap: onTap,
        child: Padding(
          padding: const EdgeInsets.all(14),
          child: Row(
            children: [
              SizedBox(
                width: 54,
                child: Text(
                  order.orderNumberDisplay ?? '—',
                  style: AppText.bodyLarge.copyWith(color: Brand.forest, fontWeight: FontWeight.w900),
                ),
              ),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(order.customerName, style: AppText.body.copyWith(fontWeight: FontWeight.w700)),
                    Text(
                      meta,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: AppText.caption.copyWith(color: metaColor),
                    ),
                  ],
                ),
              ),
              const SizedBox(width: 8),
              Column(
                crossAxisAlignment: CrossAxisAlignment.end,
                children: [
                  Pill(badge.label, color: badge.color, background: badge.background),
                  const SizedBox(height: 6),
                  Text(
                    order.total == null ? '—' : rupiah(order.total!),
                    style: AppText.body.copyWith(color: Brand.forest, fontWeight: FontWeight.w900),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}
