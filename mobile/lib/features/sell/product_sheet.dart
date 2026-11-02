import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../app/theme.dart';
import '../../app/widgets.dart';
import '../../core/format/money.dart';
import '../../data/cart/cart_notifier.dart';
import '../../domain/product.dart';

/// Opens [ProductSheet]: a bottom sheet on phones, a centred dialog on tablets.
Future<void> showProductSheet(BuildContext context, Product product) => isWide(context)
    ? showDialog(
        context: context,
        useRootNavigator: true,
        builder: (_) => Dialog(
          child: ConstrainedBox(constraints: const BoxConstraints(maxWidth: 480), child: ProductSheet(product)),
        ),
      )
    : showModalBottomSheet(
        context: context,
        useRootNavigator: true, // above the tab bar
        isScrollControlled: true,
        useSafeArea: true,
        builder: (_) => ProductSheet(product),
      );

/// Photo, price, stock, variant and quantity, then "Add to order". Used by product taps and scans.
class ProductSheet extends ConsumerStatefulWidget {
  const ProductSheet(this.product, {super.key});

  final Product product;

  @override
  ConsumerState<ProductSheet> createState() => _ProductSheetState();
}

class _ProductSheetState extends ConsumerState<ProductSheet> {
  late ProductVariant? _variant = widget.product.variants.firstOrNull;
  int _qty = 1;

  void _add() {
    final product = widget.product;
    final messenger = ScaffoldMessenger.of(context);
    ref.read(cartProvider.notifier).add(product.id, variantId: _variant?.id, qty: _qty);
    Navigator.pop(context);
    messenger.showSnackBar(SnackBar(content: Text('Added: ${product.name}')));
  }

  @override
  Widget build(BuildContext context) {
    final product = widget.product;
    final inOrder = ref.watch(cartProvider.select((c) => c.qtyOf(product.id)));
    final price = _variant?.price ?? product.price;

    return SingleChildScrollView(
      padding: const EdgeInsets.fromLTRB(20, 8, 20, 20),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          ClipRRect(
            borderRadius: BorderRadius.circular(18),
            child: SizedBox(
              height: 200,
              width: double.infinity,
              child: Stack(
                fit: StackFit.expand,
                children: [
                  ProductImage(product.imageUrl),
                  Positioned(top: 12, left: 12, child: StockBadge(product.remainingStock)),
                  if (inOrder > 0)
                    Positioned(
                      bottom: 12,
                      right: 12,
                      child: Pill('$inOrder in order', color: Colors.white, background: Brand.forest),
                    ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 16),
          if (product.brand != null) Text(product.brand!, style: AppText.caption),
          Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Expanded(child: Text(product.name, style: AppText.heading)),
              const SizedBox(width: 12),
              Text(rupiah(price), style: AppText.heading.copyWith(fontWeight: FontWeight.w900)),
            ],
          ),
          Text(product.sku, style: AppText.caption),
          if (product.variants.isNotEmpty) ...[
            const SizedBox(height: 20),
            const Text('Variant', style: AppText.label),
            const SizedBox(height: 8),
            Wrap(
              spacing: 8,
              runSpacing: 8,
              children: [
                for (final v in product.variants)
                  ChoiceChip(
                    label: Text(v.name),
                    selected: v.id == _variant?.id,
                    onSelected: (_) => setState(() => _variant = v),
                    selectedColor: Brand.greenSoft,
                    labelStyle: AppText.label.copyWith(color: Brand.forest),
                    side: BorderSide(
                      color: v.id == _variant?.id ? Brand.green : Brand.divider,
                      width: v.id == _variant?.id ? 2 : 1,
                    ),
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                  ),
              ],
            ),
          ],
          const SizedBox(height: 20),
          const Text('Quantity', style: AppText.label),
          const SizedBox(height: 8),
          Row(
            children: [
              QtyStepper(
                qty: _qty,
                onMinus: _qty > 1 ? () => setState(() => _qty--) : null,
                onPlus: _qty < product.remainingStock ? () => setState(() => _qty++) : null,
              ),
              const SizedBox(width: 12),
              Text(product.outOfStock ? 'Unavailable' : 'max ${product.remainingStock}', style: AppText.caption),
            ],
          ),
          const SizedBox(height: 24),
          BrandButton(
            text: product.outOfStock ? 'Sold out' : 'Add to order · ${rupiah(price * _qty)}',
            onPressed: product.outOfStock ? null : _add,
          ),
        ],
      ),
    );
  }
}
