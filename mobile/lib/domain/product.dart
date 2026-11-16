/// A sellable shade/size. [price] is the charged price.
class ProductVariant {
  const ProductVariant({required this.id, required this.name, required this.price});

  factory ProductVariant.fromJson(Map<String, dynamic> json) =>
      ProductVariant(id: json['id'], name: json['name'], price: (json['price'] as num? ?? 0).toDouble());

  final String id;
  final String name;
  final double price;
}

class Product {
  const Product({
    required this.id,
    required this.sku,
    required this.name,
    this.brand,
    required this.price,
    this.imageUrl,
    required this.barcodes,
    required this.remainingStock,
    this.variants = const [],
    this.stockTracked = false,
    this.variantStock = const {},
  });

  final String id;
  final String sku;
  final String name;
  final String? brand;
  final double price;
  final String? imageUrl;
  final List<String> barcodes;
  final int remainingStock;
  final List<ProductVariant> variants;

  /// 6c: false = the location does not track stock, so there are no limits at all.
  final bool stockTracked;

  /// Available per variant id, for products with options.
  final Map<String, int> variantStock;

  /// What may still be sold of [variantId] (null = the product itself); null = no limit.
  int? available(String? variantId) =>
      !stockTracked ? null : (variantId == null ? remainingStock : variantStock[variantId] ?? 0);

  /// The number on the product badge: the product's, or its options' total; null = untracked.
  int? get shownStock => !stockTracked
      ? null
      : variants.isEmpty
      ? remainingStock
      : variants.fold<int>(0, (sum, v) => sum + (variantStock[v.id] ?? 0).clamp(0, 1 << 30));

  bool get outOfStock => (shownStock ?? 1) <= 0;
  bool get lowStock => (shownStock ?? 11) > 0 && (shownStock ?? 11) <= 10;
}
