import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../app/status.dart';
import '../../app/theme.dart';
import '../../app/widgets.dart';
import '../../core/format/money.dart';
import '../../core/network/api_failure.dart';
import '../../core/printer/sunmi.dart';
import '../../core/secure/token_storage.dart';
import '../../data/cart/cart_notifier.dart';
import '../../data/event/event_repository.dart';
import '../../data/product/product_repository.dart';
import '../../data/session/session.dart';
import '../../domain/product.dart';
import '../cart/order_panel.dart';
import 'product_sheet.dart';

/// Where the cashier works. Tablets: products beside the always-open order panel.
/// Phones: products with a cart bar floating above the tabs.
class SellScreen extends ConsumerStatefulWidget {
  const SellScreen({super.key});

  @override
  ConsumerState<SellScreen> createState() => _SellScreenState();
}

class _SellScreenState extends ConsumerState<SellScreen> {
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
        : showProductSheet(context, product);
  }

  @override
  Widget build(BuildContext context) {
    ref.listen(sunmiScansProvider, (_, scan) => scan.whenData(_onHardwareScan));
    final products = ref.watch(productsProvider).value ?? const <Product>[];
    final cart = ref.watch(cartProvider);
    final wide = isWide(context);
    final counts = <String, int>{'All': products.length};
    for (final p in products) {
      if (p.brand != null) counts[p.brand!] = (counts[p.brand!] ?? 0) + 1;
    }
    final shown = products.where(_matches).toList();
    final showBar = !wide && cart.count > 0;

    final browse = Column(
      children: [
        _Header(wide: wide, onSearch: (v) => setState(() => _query = v.trim())),
        SizedBox(
          height: 58,
          child: ListView(
            scrollDirection: Axis.horizontal,
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
            children: [
              for (final brand in ['All', ...counts.keys.skip(1).toList()..sort()])
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 4),
                  child: CountChip(
                    label: brand,
                    count: counts[brand]!,
                    selected: brand == _brand,
                    onTap: () => setState(() => _brand = brand),
                  ),
                ),
            ],
          ),
        ),
        Expanded(
          child: RefreshIndicator(
            onRefresh: _refresh,
            child: shown.isEmpty
                ? ListView(children: const [EmptyState('No products match.', icon: Icons.search_off)])
                : GridView.builder(
                    padding: EdgeInsets.fromLTRB(16, 4, 16, showBar ? 96 : 16),
                    physics: const AlwaysScrollableScrollPhysics(),
                    gridDelegate: const SliverGridDelegateWithMaxCrossAxisExtent(
                      maxCrossAxisExtent: 220, // 2 columns on phones, 4 beside the tablet order panel
                      mainAxisExtent: 270,
                      mainAxisSpacing: 12,
                      crossAxisSpacing: 12,
                    ),
                    itemCount: shown.length,
                    itemBuilder: (_, i) => _ProductCard(shown[i], inOrder: cart.qtyOf(shown[i].id)),
                  ),
          ),
        ),
      ],
    );

    return Scaffold(
      body: SafeArea(
        bottom: false,
        child: wide
            ? Row(
                children: [
                  Expanded(child: browse),
                  Container(
                    width: 372,
                    margin: const EdgeInsets.fromLTRB(0, 12, 12, 12),
                    padding: const EdgeInsets.all(20),
                    decoration: BoxDecoration(
                      color: Brand.surface,
                      borderRadius: BorderRadius.circular(24),
                      border: Border.all(color: Brand.divider),
                      boxShadow: AppShadows.panel,
                    ),
                    child: const OrderPanel(),
                  ),
                ],
              )
            : Stack(
                children: [
                  browse,
                  if (showBar) const Positioned(left: 12, right: 12, bottom: 12, child: _CartBar()),
                ],
              ),
      ),
    );
  }
}

/// Phone: greeting, event and sync on top, search below. Tablet: one row with the cashier on the right.
class _Header extends ConsumerWidget {
  const _Header({required this.wide, required this.onSearch});

  final bool wide;
  final ValueChanged<String> onSearch;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final name = ref.watch(tokenStorageProvider).userName ?? 'Cashier';
    final event = ref.watch(sessionEventProvider)?.name ?? '';
    const avatar = UserAvatar(radius: 21);
    final search = Expanded(
      child: TextField(
        onChanged: onSearch,
        decoration: const InputDecoration(
          hintText: 'Search name, SKU or barcode',
          prefixIcon: Icon(Icons.search, color: Brand.muted),
        ),
      ),
    );
    final scan = IconButton.filled(
      onPressed: () => context.push('/scan'),
      tooltip: 'Scan barcode',
      icon: const Icon(Icons.qr_code_scanner),
      style: IconButton.styleFrom(
        fixedSize: const Size(52, 52),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
      ),
    );

    if (wide) {
      // The cashier's name only fits beside a roomy search box (landscape tablets).
      return LayoutBuilder(
        builder: (_, box) => Padding(
          padding: const EdgeInsets.fromLTRB(16, 12, 16, 4),
          child: Row(
            children: [
              search,
              const SizedBox(width: 10),
              scan,
              const SizedBox(width: 12),
              const SyncChip(),
              const SizedBox(width: 12),
              if (box.maxWidth >= 760) ...[
                Column(
                  crossAxisAlignment: CrossAxisAlignment.end,
                  children: [
                    Text(name, style: AppText.label.copyWith(color: Brand.forest)),
                    Text(event, style: AppText.caption.copyWith(color: Brand.slate)),
                  ],
                ),
                const SizedBox(width: 10),
              ],
              avatar,
            ],
          ),
        ),
      );
    }
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 12, 16, 4),
      child: Column(
        children: [
          Row(
            children: [
              avatar,
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'Hi, ${name.split(' ').first}',
                      style: AppText.bodyLarge.copyWith(color: Brand.forest, fontWeight: FontWeight.w900),
                    ),
                    Text(
                      event,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: AppText.caption.copyWith(color: Brand.slate),
                    ),
                  ],
                ),
              ),
              const SyncChip(),
            ],
          ),
          const SizedBox(height: 12),
          Row(children: [search, const SizedBox(width: 8), scan]),
        ],
      ),
    );
  }
}

/// Photo with stock and "in order" badges, name, price, and + (adds one; products with variants open the sheet).
class _ProductCard extends ConsumerWidget {
  const _ProductCard(this.product, {required this.inOrder});

  final Product product;
  final int inOrder;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final soldOut = product.outOfStock;
    void add() => product.variants.isNotEmpty
        ? showProductSheet(context, product)
        : ref.read(cartProvider.notifier).add(product.id);

    return DecoratedBox(
      decoration: BoxDecoration(borderRadius: BorderRadius.circular(18), boxShadow: AppShadows.card),
      child: Material(
        color: Brand.surface,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(18),
          side: const BorderSide(color: Brand.cardLine),
        ),
        clipBehavior: Clip.antiAlias,
        child: InkWell(
          onTap: () => showProductSheet(context, product),
          child: Padding(
            padding: const EdgeInsets.all(8),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Expanded(
                  child: ClipRRect(
                    borderRadius: BorderRadius.circular(13),
                    child: Stack(
                      fit: StackFit.expand,
                      children: [
                        Opacity(opacity: soldOut ? 0.45 : 1, child: ProductImage(product.imageUrl)),
                        Positioned(top: 8, left: 8, child: StockBadge(product.remainingStock)),
                        if (inOrder > 0)
                          Positioned(
                            bottom: 8,
                            right: 8,
                            child: Pill('$inOrder in order', color: Colors.white, background: Brand.forest),
                          ),
                      ],
                    ),
                  ),
                ),
                const SizedBox(height: 8),
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 4),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(product.brand ?? ' ', style: AppText.caption),
                      SizedBox(
                        height: 40,
                        child: Text(
                          product.name,
                          maxLines: 2,
                          overflow: TextOverflow.ellipsis,
                          style: AppText.body.copyWith(fontWeight: FontWeight.w700, height: 1.3),
                        ),
                      ),
                    ],
                  ),
                ),
                Row(
                  children: [
                    const SizedBox(width: 4),
                    Expanded(
                      child: Text(
                        rupiah(product.price),
                        style: AppText.bodyLarge.copyWith(color: Brand.forest, fontWeight: FontWeight.w900),
                      ),
                    ),
                    IconButton.filled(
                      onPressed: soldOut ? null : add,
                      tooltip: 'Add ${product.name}',
                      icon: const Icon(Icons.add),
                      style: IconButton.styleFrom(
                        fixedSize: const Size(44, 44),
                        backgroundColor: inOrder > 0 ? Brand.greenSoft : Brand.green,
                        foregroundColor: inOrder > 0 ? Brand.forest : Colors.white,
                        disabledBackgroundColor: Brand.sunken,
                        disabledForegroundColor: Brand.faint,
                        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

/// Phone only: count, total and customer; tap to review the order.
class _CartBar extends ConsumerWidget {
  const _CartBar();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final cart = ref.watch(cartProvider);
    final priced = ref.watch(pricedCartProvider);
    final detail = [cart.customer?.name ?? 'No customer yet', if (priced.discount.toDouble() > 0) 'promo applied'];
    return Material(
      color: Brand.forest,
      elevation: 8,
      shadowColor: Brand.forest,
      borderRadius: BorderRadius.circular(20),
      child: InkWell(
        borderRadius: BorderRadius.circular(20),
        onTap: () => context.push('/cart'),
        child: Padding(
          padding: const EdgeInsets.fromLTRB(14, 8, 8, 8),
          child: Row(
            children: [
              CircleAvatar(
                radius: 19,
                backgroundColor: Brand.gold,
                child: Text(
                  '${cart.count}',
                  style: AppText.label.copyWith(color: Brand.onGold, fontWeight: FontWeight.w900),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      rupiah(priced.total.toDouble()),
                      style: AppText.bodyLarge.copyWith(color: Colors.white, fontWeight: FontWeight.w900),
                    ),
                    Text(
                      detail.join(' · '),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: AppText.caption.copyWith(color: Brand.greenTint),
                    ),
                  ],
                ),
              ),
              IgnorePointer(
                child: FilledButton(
                  onPressed: () {},
                  style: FilledButton.styleFrom(minimumSize: const Size(0, 48)),
                  child: const Text('Review'),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
