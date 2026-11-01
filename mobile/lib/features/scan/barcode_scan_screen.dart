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
import 'scanner_view.dart';

class BarcodeScanScreen extends ConsumerWidget {
  const BarcodeScanScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final cartCount = ref.watch(cartProvider.select((c) => c.count));

    Future<void> resolve(String code) async {
      final product = await ref.read(productRepositoryProvider).findByCode(ref.read(sessionEventProvider)!.id, code);
      if (!context.mounted) return;
      if (product == null) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('No product for "$code"')));
        await Future.delayed(const Duration(milliseconds: 1600));
        return;
      }
      await showModalBottomSheet(context: context, builder: (_) => ProductSheet(product));
    }

    return ScannerView(
      title: 'Scan Product',
      hint: 'Point the camera at a product barcode or QR code',
      manualLabel: 'Enter barcode / SKU',
      onCode: resolve,
      actions: [
        if (cartCount > 0)
          IconButton(
            onPressed: () => context.go('/cart'),
            icon: Badge(
              label: Text('$cartCount'),
              backgroundColor: Brand.teal,
              child: const Icon(Icons.shopping_cart_outlined, color: Colors.white),
            ),
          ),
      ],
    );
  }
}

/// A scanned product with Add to Cart; also used for hardware-scanner reads.
class ProductSheet extends ConsumerWidget {
  const ProductSheet(this.product, {super.key});

  final Product product;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final stock = stockStatus(product.remainingStock);
    return SafeArea(
      child: Padding(
        padding: const EdgeInsets.fromLTRB(20, 20, 20, 16),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                ClipRRect(
                  borderRadius: BorderRadius.circular(12),
                  child: SizedBox.square(dimension: 72, child: ProductImage(product.imageUrl)),
                ),
                const SizedBox(width: 14),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      if (product.brand != null)
                        Text(product.brand!, style: const TextStyle(color: Brand.textFaint, fontSize: 12)),
                      Text(
                        product.name,
                        style: const TextStyle(color: Brand.heading, fontSize: 15, fontWeight: FontWeight.bold),
                      ),
                      Text(
                        rupiah(product.price),
                        style: const TextStyle(color: Brand.teal, fontSize: 16, fontWeight: FontWeight.bold),
                      ),
                    ],
                  ),
                ),
              ],
            ),
            const SizedBox(height: 12),
            Text(stock.label, style: TextStyle(color: stock.color, fontSize: 13)),
            const SizedBox(height: 16),
            BrandButton(
              text: product.outOfStock ? 'Out of stock' : 'Add to Cart',
              onPressed: product.outOfStock
                  ? null
                  : () {
                      final messenger = ScaffoldMessenger.of(context);
                      ref.read(cartProvider.notifier).add(product.id);
                      Navigator.pop(context);
                      messenger.showSnackBar(SnackBar(content: Text('Added: ${product.name}')));
                    },
            ),
          ],
        ),
      ),
    );
  }
}
