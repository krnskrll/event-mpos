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
            style: TextButton.styleFrom(foregroundColor: Brand.errorInk),
            child: const Text('Cancel order'),
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
      body: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          ScreenHeader(title: "Today's orders", subtitle: '${orders.length} this session'),
          SizedBox(
            height: 56,
            child: ListView(
              scrollDirection: Axis.horizontal,
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
              children: [
                for (final (label, state) in _filters)
                  Padding(
                    padding: const EdgeInsets.symmetric(horizontal: 4),
                    child: ChoiceChip(
                      label: Text('$label  ${orders.where((o) => state == null || o.state == state).length}'),
                      selected: _filter == state,
                      onSelected: (_) => setState(() => _filter = state),
                    ),
                  ),
              ],
            ),
          ),
          if (_message != null)
            Padding(
              padding: const EdgeInsets.fromLTRB(20, 4, 20, 0),
              child: Text(_message!, style: AppText.caption.copyWith(color: Brand.slate)),
            ),
          Expanded(
            child: shown.isEmpty
                ? const EmptyState('No orders yet. Start selling!', icon: Icons.receipt_long_outlined)
                : ListView.separated(
                    padding: const EdgeInsets.fromLTRB(16, 8, 16, 16),
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
    final badge = orderBadge(order.state);
    final time =
        '${order.createdAt.hour.toString().padLeft(2, '0')}:${order.createdAt.minute.toString().padLeft(2, '0')}';
    return Card(
      clipBehavior: Clip.antiAlias,
      child: InkWell(
        // Every numbered order has a receipt on the device, including ones still sending.
        onTap: () => context.push('/receipt/${order.key}'),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        Text(
                          order.orderNumberDisplay ?? '—',
                          style: AppText.heading.copyWith(fontWeight: FontWeight.w900),
                        ),
                        const SizedBox(width: 8),
                        Text(time, style: AppText.caption),
                      ],
                    ),
                    Text(order.customerName, style: AppText.body),
                    if (order.referenceId != null)
                      Text(order.referenceId!, maxLines: 1, overflow: TextOverflow.ellipsis, style: AppText.caption),
                    if (order.error != null) Text(order.error!, style: AppText.caption.copyWith(color: Brand.errorInk)),
                    if (order.canCancel)
                      TextButton(
                        onPressed: onCancel,
                        style: TextButton.styleFrom(foregroundColor: Brand.errorInk, padding: EdgeInsets.zero),
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
                      padding: const EdgeInsets.only(top: 8),
                      child: Text(rupiah(order.total!), style: AppText.bodyLarge.copyWith(color: Brand.forest)),
                    ),
                  const SizedBox(height: 4),
                  Text(
                    'Receipt ›',
                    style: AppText.caption.copyWith(color: Brand.green, fontWeight: FontWeight.w700),
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
