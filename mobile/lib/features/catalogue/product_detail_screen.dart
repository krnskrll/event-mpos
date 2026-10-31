import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../app/theme.dart';
import '../../app/widgets.dart';
import '../../core/format/money.dart';
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

  @override
  Widget build(BuildContext context) {
    final product = ref.watch(_productProvider(widget.productId)).value;
    // Default to the first shade once the product loads.
    final variant = _variant ?? product?.variants.firstOrNull;

    return Scaffold(
      backgroundColor: Colors.white,
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
                  child: Align(
                    alignment: Alignment.topLeft,
                    child: Padding(
                      padding: const EdgeInsets.all(12),
                      child: IconButton.filled(
                        onPressed: context.pop,
                        icon: const Icon(Icons.arrow_back, color: Brand.navy),
                        style: IconButton.styleFrom(backgroundColor: Colors.white.withValues(alpha: 0.9)),
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ),
          if (product != null)
            Expanded(
              child: _Details(product, variant: variant, onVariant: (v) => setState(() => _variant = v)),
            ),
        ],
      ),
    );
  }
}

class _Details extends StatelessWidget {
  const _Details(this.product, {required this.variant, required this.onVariant});

  final Product product;
  final ProductVariant? variant;
  final ValueChanged<ProductVariant> onVariant;

  @override
  Widget build(BuildContext context) {
    final stock = stockStatus(product.remainingStock);
    final stockText = switch (product.remainingStock) {
      <= 0 => 'Out of stock',
      <= 10 => 'Low stock: ${product.remainingStock} units',
      final n => 'In stock: $n units',
    };
    const label = TextStyle(color: Color(0xFF374151), fontSize: 13, fontWeight: FontWeight.w600);

    return ListView(
      padding: const EdgeInsets.all(20),
      children: [
        if (product.brand != null) Text(product.brand!, style: const TextStyle(color: Brand.textFaint, fontSize: 13)),
        Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Expanded(
              child: Text(
                product.name,
                style: const TextStyle(color: Brand.heading, fontSize: 20, fontWeight: FontWeight.bold),
              ),
            ),
            const SizedBox(width: 8),
            Text(
              rupiah(variant?.price ?? product.price),
              style: const TextStyle(color: Brand.teal, fontSize: 20, fontWeight: FontWeight.bold),
            ),
          ],
        ),
        const SizedBox(height: 10),
        Row(
          children: [
            Icon(Icons.circle, size: 8, color: stock.color),
            const SizedBox(width: 6),
            Text(
              stockText,
              style: TextStyle(color: stock.color, fontSize: 13, fontWeight: FontWeight.w500),
            ),
          ],
        ),
        const SizedBox(height: 8),
        Text('SKU: ${product.sku}', style: const TextStyle(color: Brand.textGhost, fontSize: 11)),
        if (product.variants.isNotEmpty) ...[
          const SizedBox(height: 20),
          const Text('Shade', style: label),
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
                  showCheckmark: false,
                  selectedColor: Brand.navy,
                  backgroundColor: Brand.navActiveBg,
                  side: BorderSide.none,
                  shape: const StadiumBorder(),
                  labelStyle: TextStyle(color: v.id == variant?.id ? Colors.white : Brand.navy, fontSize: 13),
                ),
            ],
          ),
        ],
        const SizedBox(height: 20),
        Card(
          margin: EdgeInsets.zero,
          elevation: 0,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(12),
            side: const BorderSide(color: Brand.cardBorder),
          ),
          child: ExpansionTile(
            title: const Text('Details', style: label),
            shape: const Border(),
            expandedCrossAxisAlignment: CrossAxisAlignment.start,
            childrenPadding: const EdgeInsets.fromLTRB(16, 0, 16, 14),
            children: [
              for (final line in [
                'Brand: ${product.brand ?? '—'}',
                'SKU: ${product.sku}',
                if (product.barcodes.isNotEmpty) 'Barcodes: ${product.barcodes.join(', ')}',
              ])
                Text(line, style: const TextStyle(color: Brand.textMuted, fontSize: 13)),
            ],
          ),
        ),
      ],
    );
  }
}
