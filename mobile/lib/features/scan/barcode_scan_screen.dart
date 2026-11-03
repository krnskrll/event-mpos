import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../app/theme.dart';
import '../../data/cart/cart_notifier.dart';
import '../../data/product/product_repository.dart';
import '../../data/session/session.dart';
import '../sell/product_sheet.dart';
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
      await showProductSheet(context, product);
    }

    return ScannerView(
      title: 'Scan product',
      hint: 'Point at the barcode. It scans by itself.',
      manualLabel: 'Type barcode or SKU',
      onCode: resolve,
      actions: [
        if (cartCount > 0)
          ActionChip(
            onPressed: () => context.push('/cart'),
            label: Text('$cartCount in order'),
            backgroundColor: Brand.gold,
            side: BorderSide.none,
            labelStyle: AppText.label.copyWith(color: Brand.onGold, fontWeight: FontWeight.w900),
          ),
      ],
    );
  }
}
