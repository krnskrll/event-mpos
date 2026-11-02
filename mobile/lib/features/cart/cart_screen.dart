import 'package:flutter/material.dart';

import 'order_panel.dart';

/// Phone: the order opens full screen from the cart bar. Tablets show the same panel beside the products.
class CartScreen extends StatelessWidget {
  const CartScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return const Scaffold(
      body: SafeArea(
        child: Padding(
          padding: EdgeInsets.fromLTRB(8, 8, 16, 16),
          child: OrderPanel(leading: BackButton()),
        ),
      ),
    );
  }
}
