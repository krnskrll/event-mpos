import 'package:flutter/material.dart';

/// The Bhoomy mark (Brand Guidelines v2.0 §02), drawn from its 64×64 SVG so no SVG package is needed.
class BhoomyLogo extends StatelessWidget {
  const BhoomyLogo({super.key, this.size = 64, this.tile = true, this.tileColor = const Color(0xFF143A24)});

  final double size;

  /// Draw the forest rounded tile behind the mark; turn off on dark backgrounds.
  final bool tile;

  /// Lighter forest on forest backgrounds, so the tile still reads.
  final Color tileColor;

  @override
  Widget build(BuildContext context) => Semantics(
    label: 'Bhoomy',
    image: true,
    child: CustomPaint(size: Size.square(size), painter: _LogoPainter(tile, tileColor)),
  );
}

class _LogoPainter extends CustomPainter {
  const _LogoPainter(this.tile, this.tileColor);

  final bool tile;
  final Color tileColor;

  @override
  void paint(Canvas canvas, Size size) {
    canvas.scale(size.width / 64);
    Paint fill(int color) => Paint()..color = Color(color);
    if (tile) {
      canvas.drawRRect(RRect.fromLTRBR(0, 0, 64, 64, const Radius.circular(18)), Paint()..color = tileColor);
    }
    final bloom = Path()
      ..moveTo(21, 19)
      ..lineTo(26.5, 25)
      ..lineTo(32, 16.5)
      ..lineTo(37.5, 25)
      ..lineTo(43, 19)
      ..lineTo(43, 27.5)
      ..cubicTo(43, 32.7, 38.1, 36.5, 32, 36.5)
      ..cubicTo(25.9, 36.5, 21, 32.7, 21, 27.5)
      ..close();
    canvas.drawPath(bloom, fill(0xFFD99A2B));
    canvas.drawRRect(RRect.fromLTRBR(30.7, 35, 33.3, 46, const Radius.circular(1.3)), fill(0xFFEDF6F0));
    final leaves = Path()
      ..moveTo(31, 40.5)
      ..cubicTo(25, 37.5, 18.5, 39, 14.5, 43.5)
      ..cubicTo(20, 46.8, 27.5, 45.8, 31, 42.5)
      ..close()
      ..moveTo(33, 40.5)
      ..cubicTo(39, 37.5, 45.5, 39, 49.5, 43.5)
      ..cubicTo(44, 46.8, 36.5, 45.8, 33, 42.5)
      ..close();
    canvas.drawPath(leaves, fill(0xFFA9D6BC));
    final ground = Path()
      ..moveTo(14, 51.5)
      ..cubicTo(22, 47.5, 42, 47.5, 50, 51.5);
    canvas.drawPath(
      ground,
      fill(0xFF3E9268)
        ..style = PaintingStyle.stroke
        ..strokeWidth = 3.2
        ..strokeCap = StrokeCap.round,
    );
  }

  @override
  bool shouldRepaint(_LogoPainter old) => old.tile != tile || old.tileColor != tileColor;
}
