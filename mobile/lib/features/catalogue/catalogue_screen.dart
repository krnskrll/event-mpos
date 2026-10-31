import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../app/theme.dart';
import '../../app/widgets.dart';
import '../../core/connectivity.dart';
import '../../core/format/money.dart';
import '../../core/network/api_failure.dart';
import '../../data/event/event_repository.dart';
import '../../data/product/product_repository.dart';
import '../../data/session/session.dart';
import '../../domain/product.dart';

final productsProvider = StreamProvider.autoDispose(
  (ref) => ref.watch(productRepositoryProvider).watchProducts(ref.watch(sessionEventProvider)!.id),
);

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

  @override
  Widget build(BuildContext context) {
    final products = ref.watch(productsProvider).value ?? const <Product>[];
    final brands = {for (final p in products) ?p.brand}.toList()..sort();
    final shown = products.where(_matches).toList();
    final online = ref.isOnline;

    return Scaffold(
      backgroundColor: Colors.white,
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
                        const Text('Current Event', style: TextStyle(color: Brand.textFaint, fontSize: 11)),
                        Text(
                          ref.watch(sessionEventProvider)?.name ?? '',
                          style: const TextStyle(color: Brand.heading, fontSize: 13, fontWeight: FontWeight.bold),
                        ),
                      ],
                    ),
                  ),
                  Text(
                    online ? '● Synced' : '● Offline',
                    style: TextStyle(
                      color: online ? Brand.teal : Brand.warn,
                      fontSize: 12,
                      fontWeight: FontWeight.w500,
                    ),
                  ),
                ],
              ),
            ),
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 8, 16, 0),
              child: TextField(
                onChanged: (v) => setState(() => _query = v.trim()),
                decoration: InputDecoration(
                  hintText: 'Search product…',
                  prefixIcon: const Icon(Icons.search, color: Brand.textFaint),
                  isDense: true,
                ),
              ),
            ),
            SizedBox(
              height: 48,
              child: ListView(
                scrollDirection: Axis.horizontal,
                padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                children: [
                  for (final brand in ['All', ...brands])
                    Padding(
                      padding: const EdgeInsets.symmetric(horizontal: 4),
                      child: ChoiceChip(
                        label: Text(brand),
                        selected: brand == _brand,
                        onSelected: (_) => setState(() => _brand = brand),
                        showCheckmark: false,
                        selectedColor: Brand.navy,
                        backgroundColor: Brand.cardBorder,
                        side: BorderSide.none,
                        shape: const StadiumBorder(),
                        labelStyle: TextStyle(color: brand == _brand ? Colors.white : Brand.textMuted, fontSize: 13),
                      ),
                    ),
                ],
              ),
            ),
            Expanded(
              child: RefreshIndicator(
                onRefresh: _refresh,
                child: GridView.builder(
                  padding: const EdgeInsets.all(16),
                  physics: const AlwaysScrollableScrollPhysics(),
                  gridDelegate: const SliverGridDelegateWithMaxCrossAxisExtent(
                    maxCrossAxisExtent: 220, // 2 columns on phones, more on tablets
                    mainAxisSpacing: 12,
                    crossAxisSpacing: 12,
                    childAspectRatio: 0.68,
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
    final stock = stockStatus(product.remainingStock);
    final plenty = !product.outOfStock && !product.lowStock;
    return Opacity(
      opacity: product.outOfStock ? 0.5 : 1,
      child: Card(
        margin: EdgeInsets.zero,
        color: Brand.card,
        elevation: 0,
        clipBehavior: Clip.antiAlias,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(16),
          side: const BorderSide(color: Brand.cardBorder),
        ),
        child: InkWell(
          onTap: product.outOfStock ? null : () => context.push('/catalogue/${product.id}'),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // The photo takes whatever height the text leaves, so long names never overflow.
              Expanded(
                child: Stack(
                  fit: StackFit.expand,
                  children: [
                    ProductImage(product.imageUrl),
                    Align(
                      alignment: Alignment.topRight,
                      child: Padding(
                        padding: const EdgeInsets.all(8),
                        child: Pill(
                          stock.label,
                          color: plenty ? Brand.textMuted : Colors.white,
                          background: plenty ? Colors.white : stock.color,
                        ),
                      ),
                    ),
                  ],
                ),
              ),
              Padding(
                padding: const EdgeInsets.all(10),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    if (product.brand != null)
                      Text(product.brand!, style: const TextStyle(color: Brand.textFaint, fontSize: 11)),
                    Text(
                      product.name,
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(color: Brand.text, fontSize: 13, fontWeight: FontWeight.w600),
                    ),
                    const SizedBox(height: 2),
                    Text(
                      rupiah(product.price),
                      style: const TextStyle(color: Brand.teal, fontSize: 13, fontWeight: FontWeight.bold),
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
