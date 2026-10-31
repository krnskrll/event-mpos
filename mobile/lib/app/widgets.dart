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
