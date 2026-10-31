import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';

import 'theme.dart';

/// Primary filled action button (navy by default, pass [Brand.teal] for the accent variant).
class BrandButton extends StatelessWidget {
  const BrandButton({
    super.key,
    required this.text,
    required this.onPressed,
    this.loading = false,
    this.color = Brand.navy,
  });

  final String text;
  final VoidCallback? onPressed;
  final bool loading;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      width: double.infinity,
      height: 52,
      child: FilledButton(
        onPressed: loading ? null : onPressed,
        style: FilledButton.styleFrom(
          backgroundColor: color,
          disabledBackgroundColor: loading ? color : Brand.fieldBorder,
          disabledForegroundColor: Brand.textFaint,
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
          textStyle: const TextStyle(fontSize: 16, fontWeight: FontWeight.w600),
        ),
        child: loading
            ? const SizedBox.square(
                dimension: 20,
                child: CircularProgressIndicator(color: Colors.white, strokeWidth: 2),
              )
            : Text(text),
      ),
    );
  }
}

/// Rounded status/label chip.
class Pill extends StatelessWidget {
  const Pill(this.text, {super.key, required this.color, required this.background});

  final String text;
  final Color color;
  final Color background;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
      decoration: BoxDecoration(color: background, borderRadius: BorderRadius.circular(999)),
      child: Text(
        text,
        style: TextStyle(color: color, fontSize: 11, fontWeight: FontWeight.bold),
      ),
    );
  }
}

/// White rounded card.
class SectionCard extends StatelessWidget {
  const SectionCard({super.key, required this.child});

  final Widget child;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(color: Brand.card, borderRadius: BorderRadius.circular(16)),
      child: child,
    );
  }
}

/// Product photo, or a neutral tile when there is none.
class ProductImage extends StatelessWidget {
  const ProductImage(this.url, {super.key});

  final String? url;

  @override
  Widget build(BuildContext context) {
    const placeholder = ColoredBox(
      color: Brand.navActiveBg,
      child: Center(child: Icon(Icons.image_outlined, color: Brand.textGhost, size: 40)),
    );
    if (url == null || url!.isEmpty) return placeholder;
    return CachedNetworkImage(imageUrl: url!, fit: BoxFit.cover, errorWidget: (_, _, _) => placeholder);
  }
}

/// Stock badge text and colours shared by the grid and the detail screen.
({String label, Color color}) stockStatus(int remaining) => switch (remaining) {
  <= 0 => (label: 'Out of stock', color: Brand.danger),
  <= 10 => (label: '$remaining left', color: Brand.warn),
  _ => (label: '$remaining left', color: Brand.ready),
};
