import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../app/theme.dart';
import '../../app/widgets.dart';
import '../../core/format/money.dart';
import '../../data/order/order_repository.dart';
import '../../data/session/session.dart';
import '../../domain/order.dart';

final ordersProvider = StreamProvider.autoDispose(
  (ref) => ref.watch(orderRepositoryProvider).watchOrders(ref.watch(sessionEventProvider)!.id),
);

const _filters = <(String, OrderState?)>[
  ('All', null),
  ('Confirmed', OrderState.confirmed),
  ('Syncing', OrderState.pending),
  ('Failed', OrderState.failed),
  ('Cancelled', OrderState.cancelled),
];

({String label, Color color, Color background}) _badge(OrderState state) => switch (state) {
  OrderState.pending => (label: 'SYNCING', color: Brand.packing, background: Brand.packingBg),
  OrderState.confirmed => (label: 'CONFIRMED', color: Brand.ready, background: Brand.readyBg),
  OrderState.failed => (label: 'FAILED', color: Brand.danger, background: Brand.collectedBg),
  OrderState.cancelled => (label: 'CANCELLED', color: Brand.collected, background: Brand.collectedBg),
};

class OrdersScreen extends ConsumerStatefulWidget {
  const OrdersScreen({super.key});

  @override
  ConsumerState<OrdersScreen> createState() => _OrdersScreenState();
}

class _OrdersScreenState extends ConsumerState<OrdersScreen> {
  OrderState? _filter;
  String? _message;

  Future<void> _cancel(OrderRecord order) async {
    final ok = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text('Cancel order ${order.orderNumberDisplay}?'),
        content: Text(
          "${order.customerName}'s order is cancelled and its stock goes back on sale. "
          'Only possible before the warehouse starts packing.',
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context, false), child: const Text('Keep')),
          TextButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text('Cancel order', style: TextStyle(color: Brand.danger)),
          ),
        ],
      ),
    );
    if (ok != true) return;
    final error = await ref.read(orderRepositoryProvider).cancel(order.key);
    setState(() => _message = error ?? 'Order cancelled — stock returned.');
  }

  @override
  Widget build(BuildContext context) {
    final orders = ref.watch(ordersProvider).value ?? const <OrderRecord>[];
    final shown = orders.where((o) => _filter == null || o.state == _filter).toList();

    return Scaffold(
      backgroundColor: Brand.bg,
      body: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            color: Colors.white,
            width: double.infinity,
            padding: EdgeInsets.fromLTRB(20, MediaQuery.paddingOf(context).top + 12, 20, 4),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text('Session summary', style: TextStyle(color: Brand.textFaint, fontSize: 12)),
                Row(
                  children: [
                    const Text(
                      "Today's Orders",
                      style: TextStyle(color: Brand.heading, fontSize: 18, fontWeight: FontWeight.bold),
                    ),
                    const SizedBox(width: 8),
                    Pill('${orders.length}', color: Brand.navy, background: Brand.navActiveBg),
                  ],
                ),
                SingleChildScrollView(
                  scrollDirection: Axis.horizontal,
                  child: Row(
                    children: [
                      for (final (label, state) in _filters)
                        TextButton(
                          onPressed: () => setState(() => _filter = state),
                          style: TextButton.styleFrom(
                            foregroundColor: _filter == state ? Brand.navy : Brand.textFaint,
                            textStyle: TextStyle(fontWeight: _filter == state ? FontWeight.bold : FontWeight.normal),
                          ),
                          child: Text('$label ${orders.where((o) => state == null || o.state == state).length}'),
                        ),
                    ],
                  ),
                ),
              ],
            ),
          ),
          if (_message != null)
            Padding(
              padding: const EdgeInsets.fromLTRB(20, 10, 20, 0),
              child: Text(_message!, style: const TextStyle(color: Brand.textMuted, fontSize: 12)),
            ),
          Expanded(
            child: shown.isEmpty
                ? const Center(
                    child: Text('No orders yet. Start selling!', style: TextStyle(color: Brand.textMuted)),
                  )
                : ListView.separated(
                    padding: const EdgeInsets.all(16),
                    itemCount: shown.length,
                    separatorBuilder: (_, _) => const SizedBox(height: 10),
                    itemBuilder: (_, i) => _OrderRow(shown[i], onCancel: () => _cancel(shown[i])),
                  ),
          ),
        ],
      ),
    );
  }
}

class _OrderRow extends StatelessWidget {
  const _OrderRow(this.order, {required this.onCancel});

  final OrderRecord order;
  final VoidCallback onCancel;

  @override
  Widget build(BuildContext context) {
    final badge = _badge(order.state);
    final time =
        '${order.createdAt.hour.toString().padLeft(2, '0')}:${order.createdAt.minute.toString().padLeft(2, '0')}';
    return Material(
      color: Brand.card,
      borderRadius: BorderRadius.circular(16),
      child: InkWell(
        borderRadius: BorderRadius.circular(16),
        // Every numbered order has a receipt on the device, including ones still syncing.
        onTap: () => context.push('/receipt/${order.key}'),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Row(
            children: [
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    order.orderNumberDisplay ?? '—',
                    style: const TextStyle(color: Brand.heading, fontSize: 20, fontWeight: FontWeight.w800),
                  ),
                  Text(order.customerName, style: const TextStyle(color: Brand.textFaint, fontSize: 11)),
                  Text(time, style: const TextStyle(color: Brand.textGhost, fontSize: 10)),
                ],
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    if (order.referenceId != null)
                      Text(
                        order.referenceId!,
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: const TextStyle(color: Brand.textMuted, fontSize: 11),
                      ),
                    if (order.error != null)
                      Text(order.error!, style: const TextStyle(color: Brand.danger, fontSize: 11)),
                    const Text(
                      'Tap for receipt ›',
                      style: TextStyle(color: Brand.teal, fontSize: 11, fontWeight: FontWeight.w600),
                    ),
                    if (order.canCancel)
                      TextButton(
                        onPressed: onCancel,
                        style: TextButton.styleFrom(foregroundColor: Brand.danger, padding: EdgeInsets.zero),
                        child: const Text('Cancel order'),
                      ),
                  ],
                ),
              ),
              Column(
                crossAxisAlignment: CrossAxisAlignment.end,
                children: [
                  Pill(badge.label, color: badge.color, background: badge.background),
                  if (order.total != null)
                    Padding(
                      padding: const EdgeInsets.only(top: 6),
                      child: Text(
                        rupiah(order.total!),
                        style: const TextStyle(color: Brand.heading, fontSize: 14, fontWeight: FontWeight.bold),
                      ),
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
