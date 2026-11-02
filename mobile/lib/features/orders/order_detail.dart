import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../app/theme.dart';
import '../../app/widgets.dart';
import '../../core/connectivity.dart';
import '../../core/format/money.dart';
import '../../data/order/order_repository.dart';
import '../../data/product/product_repository.dart';
import '../../domain/order.dart';
import '../../domain/product.dart';
import '../receipt/receipt_paper.dart';
import 'orders_screen.dart';

/// Phone: one order full screen. Tablets show [OrderDetail] beside the list instead.
class OrderDetailScreen extends ConsumerWidget {
  const OrderDetailScreen(this.orderKey, {super.key});

  final String orderKey;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final order = (ref.watch(ordersProvider).value ?? const <OrderRecord>[])
        .where((o) => o.key == orderKey)
        .firstOrNull;
    return Scaffold(
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.fromLTRB(8, 8, 16, 16),
          child: order == null
              ? const EmptyState('Order not found.', icon: Icons.receipt_long_outlined)
              : OrderDetail(order, leading: const BackButton()),
        ),
      ),
    );
  }
}

/// Number and status, a Details / Receipt switch, and Reprint / Cancel.
class OrderDetail extends ConsumerStatefulWidget {
  const OrderDetail(this.order, {super.key, this.leading});

  final OrderRecord order;
  final Widget? leading;

  @override
  ConsumerState<OrderDetail> createState() => _OrderDetailState();
}

class _OrderDetailState extends ConsumerState<OrderDetail> {
  bool _receipt = false;

  @override
  void didUpdateWidget(OrderDetail old) {
    super.didUpdateWidget(old);
    if (old.order.key != widget.order.key) _receipt = false; // a newly selected order opens on Details
  }

  Future<void> _cancel() async {
    final order = widget.order;
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
    if (ok != true || !mounted) return;
    final messenger = ScaffoldMessenger.of(context);
    final error = await ref.read(orderRepositoryProvider).cancel(order.key);
    ref.invalidate(receiptLinesProvider(order.key));
    messenger.showSnackBar(SnackBar(content: Text(error ?? 'Order cancelled, stock returned')));
  }

  @override
  Widget build(BuildContext context) {
    final order = widget.order;
    final badge = orderBadge(order.state);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Row(
          children: [
            ?widget.leading,
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(order.orderNumberDisplay ?? '—', style: AppText.title.copyWith(fontSize: 28)),
                  Text(
                    '${TimeOfDay.fromDateTime(order.createdAt).format(context)} · ${_items(order.itemCount)}',
                    style: AppText.caption.copyWith(color: Brand.slate),
                  ),
                ],
              ),
            ),
            Pill(badge.label, color: badge.color, background: badge.background),
          ],
        ),
        const SizedBox(height: 14),
        _ViewSwitch(receipt: _receipt, onChanged: (v) => setState(() => _receipt = v)),
        const SizedBox(height: 14),
        Expanded(child: _receipt ? _ReceiptView(order.key) : _Details(order)),
        const SizedBox(height: 12),
        Row(
          children: [
            Expanded(
              child: OutlinedButton.icon(
                onPressed: () => reprintReceipt(context, ref, order.key),
                icon: const Icon(Icons.print_outlined, size: 18),
                label: const Text('Reprint'),
                style: OutlinedButton.styleFrom(minimumSize: const Size(0, 52)),
              ),
            ),
            if (order.canCancel) ...[
              const SizedBox(width: 8),
              FilledButton(
                onPressed: _cancel,
                style: FilledButton.styleFrom(
                  minimumSize: const Size(0, 52),
                  backgroundColor: Brand.errorSoft,
                  foregroundColor: Brand.errorInk,
                ),
                child: const Text('Cancel order'),
              ),
            ],
          ],
        ),
      ],
    );
  }
}

String _items(int n) => n == 1 ? '1 item' : '$n items';

class _ViewSwitch extends StatelessWidget {
  const _ViewSwitch({required this.receipt, required this.onChanged});

  final bool receipt;
  final ValueChanged<bool> onChanged;

  @override
  Widget build(BuildContext context) {
    Widget tab(String label, bool selected, bool value) => Expanded(
      child: Semantics(
        selected: selected,
        button: true,
        child: Material(
          color: selected ? Brand.surface : Colors.transparent,
          elevation: selected ? 1 : 0,
          shadowColor: Brand.forest.withValues(alpha: 0.2),
          borderRadius: BorderRadius.circular(11),
          child: InkWell(
            borderRadius: BorderRadius.circular(11),
            onTap: () => onChanged(value),
            child: SizedBox(
              height: 42,
              child: Center(
                child: Text(label, style: AppText.label.copyWith(color: selected ? Brand.forest : Brand.slate)),
              ),
            ),
          ),
        ),
      ),
    );
    return Container(
      padding: const EdgeInsets.all(4),
      decoration: BoxDecoration(color: Brand.sunken, borderRadius: BorderRadius.circular(14)),
      child: Row(children: [tab('Details', !receipt, false), const SizedBox(width: 4), tab('Receipt', receipt, true)]),
    );
  }
}

class _ReceiptView extends ConsumerWidget {
  const _ReceiptView(this.orderKey);

  final String orderKey;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final lines = ref.watch(receiptLinesProvider(orderKey)).value;
    return DecoratedBox(
      decoration: BoxDecoration(color: Brand.canvas, borderRadius: BorderRadius.circular(16)),
      child: lines == null
          ? const EmptyState('No receipt stored for this order.', icon: Icons.receipt_outlined)
          : SingleChildScrollView(padding: const EdgeInsets.all(14), child: ReceiptPaper(lines)),
    );
  }
}

/// Customer, items, totals and what has happened to the order so far.
class _Details extends ConsumerWidget {
  const _Details(this.order);

  final OrderRecord order;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final products = {for (final p in ref.watch(productsProvider).value ?? const <Product>[]) p.id: p};
    final lines = [
      for (final l in order.lines)
        (
          line: l,
          product: products[l.productId],
          variant: products[l.productId]?.variants.where((v) => v.id == l.variantId).firstOrNull,
        ),
    ];
    // Older orders (before database v3) have no saved subtotal: rebuild it from today's prices.
    final double subtotal =
        order.subtotal ??
        lines.fold<double>(0, (sum, l) => sum + (l.variant?.price ?? l.product?.price ?? 0) * l.line.qty);
    final double total = order.total ?? subtotal;
    final double discount = order.discount ?? (total < subtotal ? subtotal - total : 0);
    final customerDetail = [
      if (order.customerTier != null) '${order.customerTier} member',
      ?order.customerQr,
    ].join(' · ');

    return ListView(
      children: [
        Container(
          padding: const EdgeInsets.all(12),
          decoration: BoxDecoration(color: Brand.greenSoft, borderRadius: BorderRadius.circular(16)),
          child: Row(
            children: [
              CircleAvatar(
                radius: 20,
                backgroundColor: Brand.green,
                child: Text(
                  initials(order.customerName),
                  style: AppText.label.copyWith(color: Colors.white, fontWeight: FontWeight.w900),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      order.customerName,
                      style: AppText.body.copyWith(fontWeight: FontWeight.w700, color: Brand.forest),
                    ),
                    if (customerDetail.isNotEmpty)
                      Text(customerDetail, style: AppText.caption.copyWith(color: Brand.forest)),
                  ],
                ),
              ),
            ],
          ),
        ),
        const SizedBox(height: 16),
        const _SectionLabel('Items'),
        for (final l in lines)
          Padding(
            padding: const EdgeInsets.symmetric(vertical: 8),
            child: Row(
              children: [
                ClipRRect(
                  borderRadius: BorderRadius.circular(11),
                  child: SizedBox.square(dimension: 44, child: ProductImage(l.product?.imageUrl)),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        l.product?.name ?? 'Product no longer listed',
                        style: AppText.body.copyWith(fontWeight: FontWeight.w700),
                      ),
                      Text(
                        [
                          ?l.variant?.name,
                          '${l.line.qty} × ${rupiah(l.variant?.price ?? l.product?.price ?? 0)}',
                        ].join(' · '),
                        style: AppText.caption,
                      ),
                    ],
                  ),
                ),
                Text(
                  rupiah((l.variant?.price ?? l.product?.price ?? 0) * l.line.qty),
                  style: AppText.body.copyWith(color: Brand.forest, fontWeight: FontWeight.w900),
                ),
              ],
            ),
          ),
        const Divider(),
        const SizedBox(height: 8),
        _Amount('Subtotal', rupiah(subtotal)),
        if (discount > 0) _Amount(order.promotionName ?? 'Discount', '−${rupiah(discount)}', color: Brand.successInk),
        const SizedBox(height: 4),
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          crossAxisAlignment: CrossAxisAlignment.baseline,
          textBaseline: TextBaseline.alphabetic,
          children: [
            Text('Total', style: AppText.bodyLarge.copyWith(color: Brand.forest)),
            Text(
              rupiah(order.total ?? subtotal - discount),
              style: AppText.heading.copyWith(fontWeight: FontWeight.w900),
            ),
          ],
        ),
        const SizedBox(height: 20),
        const _SectionLabel('Progress'),
        const SizedBox(height: 6),
        ..._steps(context, ref, order),
      ],
    );
  }
}

class _SectionLabel extends StatelessWidget {
  const _SectionLabel(this.text);

  final String text;

  @override
  Widget build(BuildContext context) => Text(
    text.toUpperCase(),
    style: AppText.caption.copyWith(color: Brand.slate, fontWeight: FontWeight.w700, letterSpacing: 0.5),
  );
}

class _Amount extends StatelessWidget {
  const _Amount(this.label, this.value, {this.color = Brand.ink});

  final String label;
  final String value;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 2),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Flexible(
            child: Text(label, style: AppText.body.copyWith(color: Brand.muted)),
          ),
          Text(
            value,
            style: AppText.body.copyWith(color: color, fontWeight: FontWeight.w700),
          ),
        ],
      ),
    );
  }
}

enum _StepKind { done, waiting, warning, failed }

/// Confirmed on the device → receipt printed → sent to the server (or why not).
List<Widget> _steps(BuildContext context, WidgetRef ref, OrderRecord order) {
  final time = TimeOfDay.fromDateTime(order.createdAt).format(context);
  final printed = ref.watch(receiptPrintedProvider(order.key)).value;
  final steps = <(String, String, _StepKind)>[
    ('Confirmed on this device', '$time · number ${order.orderNumberDisplay ?? '—'}', _StepKind.done),
    if (printed != null)
      printed
          ? ('Receipt printed', time, _StepKind.done)
          : ('Receipt not printed', 'Check the printer, then Reprint', _StepKind.warning),
    switch (order.state) {
      OrderState.pending => (
        'Sending to server',
        ref.isOnline ? 'Uploading now' : 'Waiting for internet · sends by itself',
        _StepKind.waiting,
      ),
      OrderState.confirmed => ('Sent to server', order.referenceId ?? '', _StepKind.done),
      OrderState.failed => ('Not accepted by the server', order.error ?? 'Rejected', _StepKind.failed),
      OrderState.cancelled => ('Cancelled', 'Stock returned', _StepKind.warning),
    },
  ];
  return [
    for (final (i, (title, subtitle, kind)) in steps.indexed)
      IntrinsicHeight(
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            SizedBox(
              width: 22,
              child: Column(
                children: [
                  _StepDot(kind),
                  if (i < steps.length - 1)
                    Expanded(
                      child: Container(
                        width: 2,
                        margin: const EdgeInsets.symmetric(vertical: 2),
                        color: Brand.greenTint,
                      ),
                    ),
                ],
              ),
            ),
            const SizedBox(width: 10),
            Expanded(
              child: Padding(
                padding: const EdgeInsets.only(bottom: 10),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      title,
                      style: AppText.label.copyWith(
                        color: switch (kind) {
                          _StepKind.failed => Brand.errorInk,
                          _StepKind.waiting => Brand.infoInk,
                          _StepKind.warning => Brand.goldInk,
                          _StepKind.done => Brand.forest,
                        },
                      ),
                    ),
                    if (subtitle.isNotEmpty) Text(subtitle, style: AppText.caption),
                  ],
                ),
              ),
            ),
          ],
        ),
      ),
  ];
}

class _StepDot extends StatelessWidget {
  const _StepDot(this.kind);

  final _StepKind kind;

  @override
  Widget build(BuildContext context) {
    final (fill, ring, icon) = switch (kind) {
      _StepKind.done => (Brand.green, Brand.green, Icons.check),
      _StepKind.waiting => (Brand.surface, Brand.info, null),
      _StepKind.warning => (Brand.goldSoft, Brand.warning, Icons.priority_high),
      _StepKind.failed => (Brand.errorSoft, Brand.error, Icons.close),
    };
    return Container(
      width: 20,
      height: 20,
      decoration: BoxDecoration(
        color: fill,
        shape: BoxShape.circle,
        border: Border.all(color: ring, width: 2),
      ),
      child: icon == null ? null : Icon(icon, size: 12, color: kind == _StepKind.done ? Colors.white : ring),
    );
  }
}
