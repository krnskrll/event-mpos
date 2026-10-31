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

  bool get outOfStock => remainingStock <= 0;
  bool get lowStock => remainingStock > 0 && remainingStock <= 10;
}
