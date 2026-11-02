import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';

import 'theme.dart';

/// Full-width main action. [secondary] gives the outlined variant for the less important choice.
class BrandButton extends StatelessWidget {
  const BrandButton({
    super.key,
    required this.text,
    required this.onPressed,
    this.loading = false,
    this.secondary = false,
  });

  final String text;
  final VoidCallback? onPressed;
  final bool loading;
  final bool secondary;

  @override
  Widget build(BuildContext context) {
    final shape = RoundedRectangleBorder(borderRadius: BorderRadius.circular(14));
    final textStyle = AppText.button.copyWith(fontSize: 16);
    final child = loading
        ? const SizedBox.square(dimension: 20, child: CircularProgressIndicator(color: Colors.white, strokeWidth: 2))
        : Text(text);
    return SizedBox(
      width: double.infinity,
      height: 56,
      child: secondary
          ? OutlinedButton(
              onPressed: onPressed,
              style: OutlinedButton.styleFrom(shape: shape, textStyle: textStyle),
              child: Text(text),
            )
          : FilledButton(
              onPressed: loading ? null : onPressed,
              style: FilledButton.styleFrom(
                shape: shape,
                textStyle: textStyle,
                disabledBackgroundColor: loading ? Brand.green : Brand.sunken,
                disabledForegroundColor: Brand.muted,
              ),
              child: child,
            ),
    );
  }
}

/// Light screen header on the canvas: title, optional subtitle line and trailing actions.
class ScreenHeader extends StatelessWidget {
  const ScreenHeader({super.key, required this.title, this.subtitle, this.actions = const []});

  final String title;
  final String? subtitle;
  final List<Widget> actions;

  @override
  Widget build(BuildContext context) {
    return SafeArea(
      bottom: false,
      child: Padding(
        padding: const EdgeInsets.fromLTRB(20, 16, 12, 12),
        child: Row(
          children: [
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(title, style: AppText.title),
                  if (subtitle != null) Text(subtitle!, style: AppText.caption.copyWith(color: Brand.slate)),
                ],
              ),
            ),
            ...actions,
          ],
        ),
      ),
    );
  }
}

/// Rounded status chip: dark text on a soft background.
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
        style: TextStyle(color: color, fontSize: 12, fontWeight: FontWeight.w700),
      ),
    );
  }
}

/// White rounded card with the warm hairline border.
class SectionCard extends StatelessWidget {
  const SectionCard({super.key, required this.child});

  final Widget child;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Brand.surface,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: Brand.cardLine),
      ),
      child: child,
    );
  }
}

/// − qty + control on a grey track: white minus, solid green plus. A null callback disables that side.
class QtyStepper extends StatelessWidget {
  const QtyStepper({super.key, required this.qty, this.onMinus, this.onPlus});

  final int qty;
  final VoidCallback? onMinus;
  final VoidCallback? onPlus;

  @override
  Widget build(BuildContext context) {
    Widget button(IconData icon, VoidCallback? onPressed, Color bg, Color fg, String label) => IconButton(
      onPressed: onPressed,
      tooltip: label,
      icon: Icon(icon, size: 18),
      style: IconButton.styleFrom(
        backgroundColor: bg,
        foregroundColor: fg,
        disabledBackgroundColor: Brand.surface,
        disabledForegroundColor: Brand.faint,
        fixedSize: const Size(40, 36),
        minimumSize: const Size(40, 36),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
      ),
    );
    return Container(
      padding: const EdgeInsets.all(3),
      decoration: BoxDecoration(color: Brand.sunken, borderRadius: BorderRadius.circular(11)),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          button(Icons.remove, onMinus, Brand.surface, Brand.ink, 'Remove one'),
          SizedBox(
            width: 30,
            child: Text(
              '$qty',
              textAlign: TextAlign.center,
              style: AppText.bodyLarge.copyWith(color: Brand.forest),
            ),
          ),
          button(Icons.add, onPlus, Brand.green, Colors.white, 'Add one'),
        ],
      ),
    );
  }
}

/// Filter chip with its count in a small pill.
class CountChip extends StatelessWidget {
  const CountChip({super.key, required this.label, required this.count, required this.selected, required this.onTap});

  final String label;
  final int count;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return ChoiceChip(
      selected: selected,
      onSelected: (_) => onTap(),
      label: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(label),
          const SizedBox(width: 8),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
            decoration: BoxDecoration(
              color: selected ? Brand.greenPressed : Brand.sunken,
              borderRadius: BorderRadius.circular(999),
            ),
            child: Text(
              '$count',
              style: AppText.caption.copyWith(
                color: selected ? Colors.white : Brand.slate,
                fontWeight: FontWeight.w700,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

/// Centred message for an empty list.
class EmptyState extends StatelessWidget {
  const EmptyState(this.message, {super.key, this.icon = Icons.inbox_outlined});

  final String message;
  final IconData icon;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(32),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(icon, size: 40, color: Brand.faint),
            const SizedBox(height: 12),
            Text(
              message,
              textAlign: TextAlign.center,
              style: AppText.body.copyWith(color: Brand.muted),
            ),
          ],
        ),
      ),
    );
  }
}

/// Product photo, or a soft tile when there is none.
class ProductImage extends StatelessWidget {
  const ProductImage(this.url, {super.key});

  final String? url;

  @override
  Widget build(BuildContext context) {
    const placeholder = ColoredBox(
      color: Brand.greenSoft,
      child: Center(child: Icon(Icons.image_outlined, color: Brand.greenTint, size: 40)),
    );
    if (url == null || url!.isEmpty) return placeholder;
    return CachedNetworkImage(imageUrl: url!, fit: BoxFit.cover, errorWidget: (_, _, _) => placeholder);
  }
}

/// Stock wording and badge colours shared by the product card and the product sheet.
({String label, Color background, Color color}) stockStatus(int remaining) => switch (remaining) {
  <= 0 => (label: 'Sold out', background: Brand.errorStrong, color: Colors.white),
  <= 10 => (label: 'Only $remaining left', background: Brand.goldTint, color: Brand.goldInk),
  _ => (label: '$remaining left', background: Brand.green, color: Colors.white),
};

/// Stock badge that sits on the product photo.
class StockBadge extends StatelessWidget {
  const StockBadge(this.remaining, {super.key});

  final int remaining;

  @override
  Widget build(BuildContext context) {
    final stock = stockStatus(remaining);
    return Pill(stock.label, color: stock.color, background: stock.background);
  }
}

/// Up to two initials for an avatar.
String initials(String name) =>
    name.split(' ').where((w) => w.isNotEmpty).take(2).map((w) => w[0].toUpperCase()).join();

/// Tablet and wide screens get the side-by-side layouts.
bool isWide(BuildContext context) => MediaQuery.sizeOf(context).width >= 840;
