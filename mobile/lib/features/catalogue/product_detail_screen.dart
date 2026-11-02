import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../app/theme.dart';
import '../../app/widgets.dart';
import '../../core/format/money.dart';
import '../../data/cart/cart_notifier.dart';
import '../../data/product/product_repository.dart';
import '../../data/session/session.dart';
import '../../domain/product.dart';

final _productProvider = FutureProvider.autoDispose.family(
  (ref, String id) => ref.watch(productRepositoryProvider).product(ref.watch(sessionEventProvider)!.id, id),
);

class ProductDetailScreen extends ConsumerStatefulWidget {
  const ProductDetailScreen(this.productId, {super.key});

  final String productId;

  @override
  ConsumerState<ProductDetailScreen> createState() => _ProductDetailScreenState();
}

class _ProductDetailScreenState extends ConsumerState<ProductDetailScreen> {
  ProductVariant? _variant;
  int _qty = 1;
  bool _added = false;

  @override
  Widget build(BuildContext context) {
    final product = ref.watch(_productProvider(widget.productId)).value;
    // Default to the first shade once the product loads.
    final variant = _variant ?? product?.variants.firstOrNull;
    final cartCount = ref.watch(cartProvider.select((c) => c.count));

    return Scaffold(
      backgroundColor: Brand.surface,
      body: Column(
        children: [
          SizedBox(
            height: 260,
            width: double.infinity,
            child: Stack(
              fit: StackFit.expand,
              children: [
                ProductImage(product?.imageUrl),
                SafeArea(
                  child: Padding(
                    padding: const EdgeInsets.all(12),
                    child: Row(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        _CircleButton(Icons.arrow_back, context.pop, label: 'Back'),
                        _CircleButton(
                          Icons.shopping_cart_outlined,
                          () => context.go('/cart'),
                          label: 'Cart',
                          badge: cartCount,
                        ),
                      ],
                    ),
                  ),
                ),
              ],
            ),
          ),
          if (product != null) ...[
            Expanded(
              child: _Details(
                product,
                variant: variant,
                onVariant: (v) => setState(() {
                  _variant = v;
                  _added = false;
                }),
                qty: _qty,
                onQty: (q) => setState(() => _qty = q),
              ),
            ),
            SafeArea(
              top: false,
              child: Padding(
                padding: const EdgeInsets.all(20),
                child: BrandButton(
                  text: product.outOfStock
                      ? 'Sold out'
                      : _added
                      ? 'Added to cart'
                      : 'Add to cart · ${rupiah((variant?.price ?? product.price) * _qty)}',
                  onPressed: product.outOfStock
                      ? null
                      : () {
                          ref.read(cartProvider.notifier).add(product.id, variantId: variant?.id, qty: _qty);
                          setState(() => _added = true);
                        },
                ),
              ),
            ),
          ],
        ],
      ),
    );
  }
}

class _CircleButton extends StatelessWidget {
  const _CircleButton(this.icon, this.onPressed, {required this.label, this.badge = 0});

  final IconData icon;
  final VoidCallback onPressed;
  final String label;
  final int badge;

  @override
  Widget build(BuildContext context) {
    return IconButton.filled(
      onPressed: onPressed,
      tooltip: label,
      icon: Badge(
        isLabelVisible: badge > 0,
        label: Text('$badge'),
        child: Icon(icon, color: Brand.forest),
      ),
      style: IconButton.styleFrom(backgroundColor: Colors.white.withValues(alpha: 0.92)),
    );
  }
}

class _Details extends StatelessWidget {
  const _Details(
    this.product, {
    required this.variant,
    required this.onVariant,
    required this.qty,
    required this.onQty,
  });

  final Product product;
  final ProductVariant? variant;
  final ValueChanged<ProductVariant> onVariant;
  final int qty;
  final ValueChanged<int> onQty;

  @override
  Widget build(BuildContext context) {
    return ListView(
      padding: const EdgeInsets.all(20),
      children: [
        if (product.brand != null) Text(product.brand!, style: AppText.caption),
        Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Expanded(child: Text(product.name, style: AppText.heading)),
            const SizedBox(width: 12),
            Text(rupiah(variant?.price ?? product.price), style: AppText.heading.copyWith(fontWeight: FontWeight.w900)),
          ],
        ),
        const SizedBox(height: 8),
        StockLabel(product.remainingStock),
        const SizedBox(height: 4),
        Text('SKU ${product.sku}', style: AppText.caption),
        if (product.variants.isNotEmpty) ...[
          const SizedBox(height: 24),
          const Text('Variant', style: AppText.label),
          const SizedBox(height: 8),
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: [
              for (final v in product.variants)
                ChoiceChip(
                  label: Text(v.name),
                  selected: v.id == variant?.id,
                  onSelected: (_) => onVariant(v),
                  selectedColor: Brand.greenSoft,
                  labelStyle: AppText.label.copyWith(color: Brand.forest),
                  side: BorderSide(
                    color: v.id == variant?.id ? Brand.green : Brand.divider,
                    width: v.id == variant?.id ? 2 : 1,
                  ),
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                ),
            ],
          ),
        ],
        const SizedBox(height: 24),
        const Text('Quantity', style: AppText.label),
        const SizedBox(height: 8),
        Row(
          children: [
            QtyStepper(
              qty: qty,
              onMinus: qty > 1 ? () => onQty(qty - 1) : null,
              onPlus: qty < product.remainingStock ? () => onQty(qty + 1) : null,
            ),
            const SizedBox(width: 12),
            Text(product.outOfStock ? 'Unavailable' : 'max ${product.remainingStock}', style: AppText.caption),
          ],
        ),
        const SizedBox(height: 24),
        Card(
          clipBehavior: Clip.antiAlias,
          child: ExpansionTile(
            title: const Text('Details', style: AppText.label),
            shape: const Border(),
            expandedCrossAxisAlignment: CrossAxisAlignment.start,
            childrenPadding: const EdgeInsets.fromLTRB(16, 0, 16, 14),
            children: [
              for (final line in [
                'Brand: ${product.brand ?? '—'}',
                'SKU: ${product.sku}',
                if (product.barcodes.isNotEmpty) 'Barcodes: ${product.barcodes.join(', ')}',
              ])
                Text(line, style: AppText.caption),
            ],
          ),
        ),
      ],
    );
  }
}
