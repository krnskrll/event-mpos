import 'package:decimal/decimal.dart';

import 'customer.dart';
import 'product.dart';
import 'promotion.dart';

/// Lines are keyed "productId|variantId" so each shade is its own line.
String cartKey(String productId, String? variantId) => '$productId|${variantId ?? ''}';

({String productId, String? variantId}) parseCartKey(String key) {
  final [productId, variantId] = key.split('|');
  return (productId: productId, variantId: variantId.isEmpty ? null : variantId);
}

/// The sale being built: quantities per line key, plus the customer.
class Cart {
  const Cart({this.items = const {}, this.customer});

  factory Cart.fromJson(Map<String, dynamic> json) => Cart(
    items: Map<String, int>.from(json['items']),
    customer: json['customer'] == null ? null : Customer.fromJson(json['customer']),
  );

  final Map<String, int> items;
  final Customer? customer;

  bool get isEmpty => items.isEmpty && customer == null;
  int get count => items.values.fold(0, (sum, qty) => sum + qty);

  Map<String, dynamic> toJson() => {'items': items, 'customer': customer?.toJson()};
}

class CartItem {
  const CartItem(this.key, this.product, this.variant, this.qty);

  final String key;
  final Product product;
  final ProductVariant? variant;
  final int qty;

  Decimal get unitPrice => Decimal.parse((variant?.price ?? product.price).toString());
  Decimal get lineTotal => unitPrice * Decimal.fromInt(qty);
  String get displayName => variant == null ? product.name : '${product.name} — ${variant!.name}';
}

/// The cart with prices and the best promotion applied (a preview; the server confirms).
class PricedCart {
  const PricedCart(this.items, this.subtotal, this.discount, this.promotionName);

  factory PricedCart.of(Cart cart, List<Product> products, List<Promotion> promotions, DateTime now) {
    final byId = {for (final p in products) p.id: p};
    final items = [
      for (final MapEntry(:key, value: qty) in cart.items.entries)
        if (byId[parseCartKey(key).productId] case final product?)
          CartItem(key, product, product.variants.where((v) => v.id == parseCartKey(key).variantId).firstOrNull, qty),
    ];
    final subtotal = items.fold(Decimal.zero, (sum, i) => sum + i.lineTotal);
    final applied = PromotionEngine.best(
      [for (final i in items) PricedLine(i.product.id, i.unitPrice, i.qty)],
      promotions,
      now,
    );
    return PricedCart(items, subtotal, applied?.discount ?? Decimal.zero, applied?.promotion.name);
  }

  final List<CartItem> items;
  final Decimal subtotal;
  final Decimal discount;
  final String? promotionName;

  Decimal get total => subtotal - discount;
}
