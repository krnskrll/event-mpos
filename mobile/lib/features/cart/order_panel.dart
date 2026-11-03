import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../app/theme.dart';
import '../../app/widgets.dart';
import '../../core/format/money.dart';
import '../../data/cart/cart_notifier.dart';
import '../../data/order/order_repository.dart';
import '../../data/session/session.dart';
import '../../domain/cart.dart';
import '../../domain/customer.dart';
import 'draft_dialogs.dart';

/// The order being built: customer, lines, promotion, total and Confirm. The tablet's side panel and
/// the phone's cart page both show this; [leading] is the phone's back button.
class OrderPanel extends ConsumerStatefulWidget {
  const OrderPanel({super.key, this.leading});

  final Widget? leading;

  @override
  ConsumerState<OrderPanel> createState() => _OrderPanelState();
}

class _OrderPanelState extends ConsumerState<OrderPanel> {
  bool _confirming = false;
  String? _error;

  /// Confirms on the device (final number + receipt), online or offline; the upload follows.
  Future<void> _confirm() async {
    setState(() {
      _confirming = true;
      _error = null;
    });
    try {
      final key = await ref
          .read(orderRepositoryProvider)
          .confirm(ref.read(sessionEventProvider)!, ref.read(cartProvider), ref.read(pricedCartProvider));
      ref.read(cartProvider.notifier).clear();
      if (mounted) context.push('/receipt/$key');
    } catch (e) {
      setState(() => _error = '$e');
    } finally {
      if (mounted) setState(() => _confirming = false);
    }
  }

  Future<void> _clear() async {
    final ok = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Clear this order?'),
        content: const Text('All items and the customer are removed. Save it as a draft first to keep it.'),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context, false), child: const Text('Keep')),
          TextButton(
            onPressed: () => Navigator.pop(context, true),
            style: TextButton.styleFrom(foregroundColor: Brand.errorInk),
            child: const Text('Clear'),
          ),
        ],
      ),
    );
    if (ok == true) ref.read(cartProvider.notifier).clear();
  }

  @override
  Widget build(BuildContext context) {
    final cart = ref.watch(cartProvider);
    final priced = ref.watch(pricedCartProvider);
    final drafts = ref.watch(draftsProvider).value?.length ?? 0;
    final notifier = ref.read(cartProvider.notifier);

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
                  Text('Current order', style: AppText.heading.copyWith(fontWeight: FontWeight.w900)),
                  Text(itemCount(cart.count), style: AppText.caption),
                ],
              ),
            ),
            FilledButton.tonal(
              onPressed: () => showDraftsDialog(context, ref),
              style: FilledButton.styleFrom(minimumSize: const Size(0, 40)),
              child: Text(drafts > 0 ? 'Drafts · $drafts' : 'Drafts'),
            ),
            if (cart.items.isNotEmpty) ...[
              const SizedBox(width: 6),
              IconButton(
                onPressed: _clear,
                tooltip: 'Clear order',
                icon: const Icon(Icons.delete_outline, size: 20),
                style: IconButton.styleFrom(backgroundColor: Brand.errorSoft, foregroundColor: Brand.errorInk),
              ),
            ],
          ],
        ),
        const SizedBox(height: 14),
        _CustomerCard(cart.customer),
        Expanded(
          child: priced.items.isEmpty
              ? const EmptyState('No items yet.\nTap + on a product to add it.', icon: Icons.shopping_basket_outlined)
              : ListView.separated(
                  padding: const EdgeInsets.symmetric(vertical: 4),
                  itemCount: priced.items.length,
                  separatorBuilder: (_, _) => const Divider(),
                  itemBuilder: (_, i) {
                    final item = priced.items[i];
                    return _OrderLine(
                      item,
                      onMinus: () => notifier.adjust(item.key, -1),
                      onPlus: item.qty < item.product.remainingStock ? () => notifier.adjust(item.key, 1) : null,
                    );
                  },
                ),
        ),
        if (priced.discount.toDouble() > 0) ...[
          _PromoBox(priced.promotionName ?? 'Discount', priced.discount.toDouble()),
          const SizedBox(height: 10),
        ],
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text('Subtotal', style: AppText.body.copyWith(color: Brand.muted)),
            Text(rupiah(priced.subtotal.toDouble()), style: AppText.body),
          ],
        ),
        const Padding(padding: EdgeInsets.symmetric(vertical: 10), child: Divider()),
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          crossAxisAlignment: CrossAxisAlignment.baseline,
          textBaseline: TextBaseline.alphabetic,
          children: [
            Text('Total', style: AppText.bodyLarge.copyWith(color: Brand.forest)),
            Text(rupiah(priced.total.toDouble()), style: AppText.total),
          ],
        ),
        if (_error != null)
          Padding(
            padding: const EdgeInsets.only(top: 8),
            child: Text(_error!, style: AppText.label.copyWith(color: Brand.errorInk)),
          ),
        const SizedBox(height: 14),
        BrandButton(
          text: cart.customer == null && cart.items.isNotEmpty ? 'Identify a customer to confirm' : 'Confirm order',
          loading: _confirming,
          onPressed: cart.customer == null || cart.items.isEmpty ? null : _confirm,
        ),
      ],
    );
  }
}

/// Linked customer on a green card, or the prompt to identify one (required before confirming).
class _CustomerCard extends StatelessWidget {
  const _CustomerCard(this.customer);

  final Customer? customer;

  @override
  Widget build(BuildContext context) {
    final c = customer;
    final detail = [if (c?.tier != null) memberLabel(c!.tier!), ?c?.qrCode].join(' · ');
    return Material(
      color: c != null ? Brand.greenSoft : Brand.surface,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(16),
        side: c != null ? BorderSide.none : const BorderSide(color: Brand.divider),
      ),
      child: InkWell(
        borderRadius: BorderRadius.circular(16),
        onTap: () => context.push('/customer'),
        child: Padding(
          padding: const EdgeInsets.all(12),
          child: Row(
            children: [
              CircleAvatar(
                radius: 21,
                backgroundColor: c != null ? Brand.green : Brand.sunken,
                child: c != null
                    ? Text(
                        initials(c.name),
                        style: AppText.label.copyWith(color: Colors.white, fontWeight: FontWeight.w900),
                      )
                    : const Icon(Icons.person_outline, color: Brand.slate),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      c?.name ?? 'Identify customer',
                      style: AppText.body.copyWith(fontWeight: FontWeight.w700, color: Brand.forest),
                    ),
                    if (c == null)
                      Text('Needed before confirming', style: AppText.caption.copyWith(color: Brand.errorInk))
                    else if (detail.isNotEmpty)
                      Text(detail, style: AppText.caption.copyWith(color: Brand.forest)),
                  ],
                ),
              ),
              Text(c == null ? 'Add' : 'Change', style: AppText.label.copyWith(color: Brand.green)),
            ],
          ),
        ),
      ),
    );
  }
}

class _OrderLine extends StatelessWidget {
  const _OrderLine(this.item, {required this.onMinus, this.onPlus});

  final CartItem item;
  final VoidCallback onMinus;
  final VoidCallback? onPlus;

  @override
  Widget build(BuildContext context) {
    final unit = rupiah(item.unitPrice.toDouble());
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 8),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          ClipRRect(
            borderRadius: BorderRadius.circular(12),
            child: SizedBox.square(dimension: 56, child: ProductImage(item.product.imageUrl)),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Expanded(
                      child: Text(item.product.name, style: AppText.body.copyWith(fontWeight: FontWeight.w700)),
                    ),
                    const SizedBox(width: 8),
                    Text(
                      rupiah(item.lineTotal.toDouble()),
                      style: AppText.body.copyWith(color: Brand.forest, fontWeight: FontWeight.w900),
                    ),
                  ],
                ),
                Text(item.variant == null ? unit : '${item.variant!.name} · $unit', style: AppText.caption),
                const SizedBox(height: 6),
                QtyStepper(qty: item.qty, onMinus: onMinus, onPlus: onPlus),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

/// The applied promotion, in gold: gold always means money saved.
class _PromoBox extends StatelessWidget {
  const _PromoBox(this.name, this.amount);

  final String name;
  final double amount;

  @override
  Widget build(BuildContext context) {
    final style = AppText.label.copyWith(color: Brand.goldInk);
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 11),
      decoration: BoxDecoration(color: Brand.goldSoft, borderRadius: BorderRadius.circular(14)),
      child: Row(
        children: [
          const Icon(Icons.sell_outlined, size: 18, color: Brand.goldInk),
          const SizedBox(width: 10),
          Expanded(child: Text(name, style: style)),
          Text('−${rupiah(amount)}', style: style.copyWith(fontWeight: FontWeight.w900)),
        ],
      ),
    );
  }
}
