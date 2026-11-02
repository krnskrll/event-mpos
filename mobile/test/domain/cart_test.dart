import 'package:flutter_test/flutter_test.dart';
import 'package:mpos/domain/cart.dart';

void main() {
  test('qtyOf adds up every variant of a product', () {
    final cart = Cart(items: {cartKey('p1', 'rose'): 2, cartKey('p1', 'nude'): 1, cartKey('p2', null): 4});
    expect([cart.qtyOf('p1'), cart.qtyOf('p2'), cart.qtyOf('p3')], [3, 4, 0]);
  });
}
