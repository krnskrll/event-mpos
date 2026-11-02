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
      body: Column(
        children: [
          ScreenHeader(
            title: 'Cart',
            subtitle: '${cart.count} items',
            actions: [
              if (cart.items.isNotEmpty)
                TextButton(onPressed: () => showSaveDraftDialog(context, ref), child: const Text('Save draft')),
              TextButton(onPressed: () => showDraftsDialog(context, ref), child: const Text('Drafts')),
            ],
          ),
          if (priced.items.isEmpty)
            const Expanded(
              child: EmptyState('Cart is empty.\nAdd products from the catalogue.', icon: Icons.shopping_cart_outlined),
            )
          else ...[
            Expanded(
              child: ListView.separated(
                padding: const EdgeInsets.fromLTRB(16, 4, 16, 16),
                itemCount: priced.items.length,
                separatorBuilder: (_, _) => const SizedBox(height: 10),
                itemBuilder: (_, i) {
                  final item = priced.items[i];
                  return SectionCard(
                    child: Row(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        ClipRRect(
                          borderRadius: BorderRadius.circular(12),
                          child: SizedBox.square(dimension: 64, child: ProductImage(item.product.imageUrl)),
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
                                    child: Text(
                                      item.displayName,
                                      maxLines: 2,
                                      style: AppText.body.copyWith(fontWeight: FontWeight.w700),
                                    ),
                                  ),
                                  const SizedBox(width: 8),
                                  Text(
                                    rupiah(item.lineTotal.toDouble()),
                                    style: AppText.body.copyWith(color: Brand.forest, fontWeight: FontWeight.w700),
                                  ),
                                ],
                              ),
                              const SizedBox(height: 8),
                              QtyStepper(
                                qty: item.qty,
                                onMinus: () => notifier.adjust(item.key, -1),
                                onPlus: () => notifier.adjust(item.key, 1),
                              ),
                            ],
                          ),
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
                        backgroundColor: customer != null ? Brand.greenSoft : Brand.sunken,
                        child: Icon(
                          customer != null ? Icons.check : Icons.person_outline,
                          color: customer != null ? Brand.green : Brand.muted,
                        ),
                      ),
                      title: Text(
                        customer == null ? 'Identify customer' : customer.name,
                        style: AppText.body.copyWith(fontWeight: FontWeight.w700, color: Brand.forest),
                      ),
                      subtitle: customer == null
                          ? Text('Needed before confirming', style: AppText.caption.copyWith(color: Brand.errorInk))
                          : customer.tier == null
                          ? null
                          : Text(
                              '${customer.tier} member',
                              style: AppText.caption.copyWith(color: Brand.goldInk, fontWeight: FontWeight.w700),
                            ),
                      trailing: Text(
                        customer == null ? 'Add' : 'Change',
                        style: AppText.label.copyWith(color: Brand.green),
                      ),
                    ),
                    const Divider(),
                    const SizedBox(height: 8),
                    _TotalRow('Subtotal', priced.subtotal.toDouble()),
                    if (priced.discount.toDouble() > 0)
                      _TotalRow(priced.promotionName ?? 'Discount', priced.discount.toDouble(), discount: true),
                    const SizedBox(height: 8),
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

class _TotalRow extends StatelessWidget {
  const _TotalRow(this.label, this.amount, {this.discount = false});

  final String label;
  final double amount;
  final bool discount;

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
            discount ? '−${rupiah(amount)}' : rupiah(amount),
            style: AppText.body.copyWith(color: discount ? Brand.successInk : Brand.ink),
          ),
        ],
      ),
    );
  }
}
