import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../app/theme.dart';
import '../../app/widgets.dart';
import '../../core/connectivity.dart';
import '../../core/format/money.dart';
import '../../core/network/api_failure.dart';
import '../../core/printer/sunmi.dart';
import '../../data/cart/cart_notifier.dart';
import '../../data/event/event_repository.dart';
import '../../data/product/product_repository.dart';
import '../../data/session/session.dart';
import '../../domain/product.dart';
import '../scan/barcode_scan_screen.dart';

class CatalogueScreen extends ConsumerStatefulWidget {
  const CatalogueScreen({super.key});

  @override
  ConsumerState<CatalogueScreen> createState() => _CatalogueScreenState();
}

class _CatalogueScreenState extends ConsumerState<CatalogueScreen> {
  String _query = '';
  String _brand = 'All';

  bool _matches(Product p) {
    final q = _query.toLowerCase();
    return (_brand == 'All' || p.brand == _brand) &&
        (p.name.toLowerCase().contains(q) || p.sku.toLowerCase().contains(q) || p.barcodes.any((b) => b.contains(q)));
  }

  /// Re-syncs the snapshot only if it changed on the server, then pulls fresh stock.
  Future<void> _refresh() async {
    final eventId = ref.read(sessionEventProvider)!.id;
    try {
      await ref.read(eventRepositoryProvider).refreshSnapshot(eventId);
      await ref.read(productRepositoryProvider).refreshStock(eventId);
    } on ApiFailure catch (e) {
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(e.message)));
    }
  }

  /// A read from the Sunmi hardware scanner opens the same sheet as a camera scan.
  Future<void> _onHardwareScan(String code) async {
    final product = await ref.read(productRepositoryProvider).findByCode(ref.read(sessionEventProvider)!.id, code);
    if (!mounted) return;
    product == null
        ? ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('No product for "$code"')))
        : showModalBottomSheet(context: context, builder: (_) => ProductSheet(product));
  }

  @override
  Widget build(BuildContext context) {
    ref.listen(sunmiScansProvider, (_, scan) => scan.whenData(_onHardwareScan));
    final customer = ref.watch(cartProvider.select((c) => c.customer));
    final products = ref.watch(productsProvider).value ?? const <Product>[];
    final counts = <String, int>{'All': products.length};
    for (final p in products) {
      if (p.brand != null) counts[p.brand!] = (counts[p.brand!] ?? 0) + 1;
    }
    final brands = counts.keys.skip(1).toList()..sort();
    final shown = products.where(_matches).toList();
    final online = ref.isOnline;

    return Scaffold(
      body: SafeArea(
        bottom: false,
        child: Column(
          children: [
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 12, 16, 4),
              child: Row(
                children: [
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text('Selling at', style: AppText.caption.copyWith(color: Brand.slate)),
                        Text(
                          ref.watch(sessionEventProvider)?.name ?? '',
                          style: AppText.label.copyWith(color: Brand.forest),
                        ),
                      ],
                    ),
                  ),
                  Icon(Icons.circle, size: 8, color: online ? Brand.success : Brand.warning),
                  const SizedBox(width: 6),
                  Text(
                    online ? 'Synced' : 'Offline',
                    style: AppText.label.copyWith(color: online ? Brand.successInk : Brand.warningInk),
                  ),
                ],
              ),
            ),
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 8, 16, 0),
              child: Row(
                children: [
                  Expanded(
                    child: TextField(
                      onChanged: (v) => setState(() => _query = v.trim()),
                      decoration: const InputDecoration(
                        hintText: 'Search name, SKU or barcode',
                        prefixIcon: Icon(Icons.search, color: Brand.muted),
                        isDense: true,
                      ),
                    ),
                  ),
                  const SizedBox(width: 8),
                  IconButton.filled(
                    onPressed: () => context.push('/scan'),
                    tooltip: 'Scan barcode',
                    icon: const Icon(Icons.qr_code_scanner),
                  ),
                  IconButton.outlined(
                    onPressed: () => context.push('/customer'),
                    tooltip: 'Identify customer',
                    icon: Icon(customer != null ? Icons.how_to_reg : Icons.person_outline),
                    style: IconButton.styleFrom(
                      foregroundColor: customer != null ? Brand.green : Brand.ink,
                      backgroundColor: customer != null ? Brand.greenSoft : Brand.surface,
                      side: BorderSide(color: customer != null ? Brand.green : Brand.fieldLine),
                    ),
                  ),
                ],
              ),
            ),
            if (customer != null)
              Container(
                width: double.infinity,
                margin: const EdgeInsets.fromLTRB(16, 10, 16, 0),
                padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                decoration: BoxDecoration(color: Brand.greenSoft, borderRadius: BorderRadius.circular(12)),
                child: Text('Customer: ${customer.name}', style: AppText.label.copyWith(color: Brand.forest)),
              ),
            SizedBox(
              height: 56,
              child: ListView(
                scrollDirection: Axis.horizontal,
                padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                children: [
                  for (final brand in ['All', ...brands])
                    Padding(
                      padding: const EdgeInsets.symmetric(horizontal: 4),
                      child: ChoiceChip(
                        label: Text('$brand  ${counts[brand]}'),
                        selected: brand == _brand,
                        onSelected: (_) => setState(() => _brand = brand),
                      ),
                    ),
                ],
              ),
            ),
            Expanded(
              child: RefreshIndicator(
                onRefresh: _refresh,
                child: GridView.builder(
                  padding: const EdgeInsets.fromLTRB(16, 4, 16, 16),
                  physics: const AlwaysScrollableScrollPhysics(),
                  gridDelegate: const SliverGridDelegateWithMaxCrossAxisExtent(
                    maxCrossAxisExtent: 220, // 2 columns on phones, more on tablets
                    mainAxisSpacing: 12,
                    crossAxisSpacing: 12,
                    childAspectRatio: 0.66,
                  ),
                  itemCount: shown.length,
                  itemBuilder: (_, i) => _ProductCard(shown[i]),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _ProductCard extends StatelessWidget {
  const _ProductCard(this.product);

  final Product product;

  @override
  Widget build(BuildContext context) {
    return Opacity(
      opacity: product.outOfStock ? 0.55 : 1,
      child: Card(
        clipBehavior: Clip.antiAlias,
        child: InkWell(
          onTap: product.outOfStock ? null : () => context.push('/product/${product.id}'),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // The photo takes whatever height the text leaves, so long names never overflow.
              Expanded(
                child: SizedBox(width: double.infinity, child: ProductImage(product.imageUrl)),
              ),
              Padding(
                padding: const EdgeInsets.fromLTRB(12, 10, 12, 12),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    if (product.brand != null) Text(product.brand!, style: AppText.caption),
                    Text(
                      product.name,
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                      style: AppText.body.copyWith(fontWeight: FontWeight.w700),
                    ),
                    const SizedBox(height: 4),
                    StockLabel(product.remainingStock),
                    const SizedBox(height: 6),
                    Text(
                      rupiah(product.price),
                      style: AppText.bodyLarge.copyWith(color: Brand.forest, fontWeight: FontWeight.w900),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
