import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../app/theme.dart';
import '../../app/widgets.dart';
import '../../core/format/money.dart';
import '../../data/cart/cart_notifier.dart';
import '../../data/order/order_repository.dart';
import '../../data/session/session.dart';
import 'draft_dialogs.dart';

class CartScreen extends ConsumerStatefulWidget {
  const CartScreen({super.key});

  @override
  ConsumerState<CartScreen> createState() => _CartScreenState();
}

class _CartScreenState extends ConsumerState<CartScreen> {
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

  @override
  Widget build(BuildContext context) {
    final cart = ref.watch(cartProvider);
    final priced = ref.watch(pricedCartProvider);
    final notifier = ref.read(cartProvider.notifier);
    final customer = cart.customer;

    return Scaffold(
      backgroundColor: Brand.bg,
      body: Column(
        children: [
          Container(
            color: Brand.navy,
            padding: EdgeInsets.fromLTRB(20, MediaQuery.paddingOf(context).top + 12, 12, 12),
            child: Row(
              children: [
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const Text(
                        'Cart',
                        style: TextStyle(color: Colors.white, fontSize: 20, fontWeight: FontWeight.bold),
                      ),
                      Text('${cart.count} items', style: const TextStyle(color: Colors.white60, fontSize: 13)),
                    ],
                  ),
                ),
                if (cart.items.isNotEmpty) _HeaderAction('Save draft', () => showSaveDraftDialog(context, ref)),
                _HeaderAction('Drafts', () => showDraftsDialog(context, ref)),
              ],
            ),
          ),
          if (priced.items.isEmpty)
            const Expanded(
              child: Center(
                child: Text(
                  'Cart is empty.\nAdd products from the catalogue.',
                  textAlign: TextAlign.center,
                  style: TextStyle(color: Brand.textMuted),
                ),
              ),
            )
          else ...[
            Expanded(
              child: ListView.separated(
                padding: const EdgeInsets.all(16),
                itemCount: priced.items.length,
                separatorBuilder: (_, _) => const SizedBox(height: 10),
                itemBuilder: (_, i) {
                  final item = priced.items[i];
                  return SectionCard(
                    child: Row(
                      children: [
                        ClipRRect(
                          borderRadius: BorderRadius.circular(10),
                          child: SizedBox.square(dimension: 52, child: ProductImage(item.product.imageUrl)),
                        ),
                        const SizedBox(width: 10),
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                item.displayName,
                                maxLines: 2,
                                style: const TextStyle(color: Brand.text, fontSize: 13, fontWeight: FontWeight.w600),
                              ),
                              Text(
                                rupiah(item.lineTotal.toDouble()),
                                style: const TextStyle(color: Brand.teal, fontSize: 13, fontWeight: FontWeight.bold),
                              ),
                            ],
                          ),
                        ),
                        IconButton.filledTonal(
                          onPressed: () => notifier.adjust(item.key, -1),
                          icon: const Icon(Icons.remove),
                        ),
                        Text('${item.qty}', style: const TextStyle(fontWeight: FontWeight.bold)),
                        IconButton.filledTonal(
                          onPressed: () => notifier.adjust(item.key, 1),
                          icon: const Icon(Icons.add),
                        ),
                      ],
                    ),
                  );
                },
              ),
            ),
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
              child: SectionCard(
                child: Column(
                  children: [
                    ListTile(
                      contentPadding: EdgeInsets.zero,
                      onTap: () => context.push('/customer'),
                      leading: CircleAvatar(
                        backgroundColor: customer != null ? Brand.tealSoft : Brand.cardBorder,
                        child: Icon(
                          customer != null ? Icons.check : Icons.person_outline,
                          color: customer != null ? Brand.teal : Brand.textMuted,
                        ),
                      ),
                      title: Text(
                        customer == null
                            ? 'Identify customer'
                            : '${customer.name}${customer.tier == null ? '' : ' · ${customer.tier}'}',
                        style: TextStyle(color: customer == null ? Brand.danger : Brand.text, fontSize: 14),
                      ),
                      trailing: Text(
                        customer == null ? 'Add' : 'Change',
                        style: const TextStyle(color: Brand.teal, fontWeight: FontWeight.w600),
                      ),
                    ),
                    const Divider(color: Brand.cardBorder),
                    _TotalRow('Subtotal', priced.subtotal.toDouble()),
                    if (priced.discount.toDouble() > 0)
                      _TotalRow('Discount (${priced.promotionName})', -priced.discount.toDouble(), color: Brand.teal),
                    _TotalRow('Total', priced.total.toDouble(), color: Brand.heading, bold: true),
                    if (_error != null)
                      Padding(
                        padding: const EdgeInsets.only(top: 8),
                        child: Text(_error!, style: const TextStyle(color: Brand.danger, fontSize: 13)),
                      ),
                    const SizedBox(height: 12),
                    BrandButton(
                      text: customer == null ? 'Identify a customer to confirm' : 'Confirm order',
                      loading: _confirming,
                      onPressed: customer == null ? null : _confirm,
                    ),
                  ],
                ),
              ),
            ),
          ],
        ],
      ),
    );
  }
}

class _HeaderAction extends StatelessWidget {
  const _HeaderAction(this.label, this.onPressed);

  final String label;
  final VoidCallback onPressed;

  @override
  Widget build(BuildContext context) {
    return TextButton(
      onPressed: onPressed,
      style: TextButton.styleFrom(foregroundColor: Colors.white),
      child: Text(label, style: const TextStyle(fontWeight: FontWeight.w600)),
    );
  }
}

class _TotalRow extends StatelessWidget {
  const _TotalRow(this.label, this.amount, {this.color = Brand.textMuted, this.bold = false});

  final String label;
  final double amount;
  final Color color;
  final bool bold;

  @override
  Widget build(BuildContext context) {
    final style = TextStyle(
      color: color,
      fontSize: bold ? 16 : 14,
      fontWeight: bold ? FontWeight.bold : FontWeight.normal,
    );
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 2),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Flexible(child: Text(label, style: style)),
          Text(rupiah(amount), style: style),
        ],
      ),
    );
  }
}
