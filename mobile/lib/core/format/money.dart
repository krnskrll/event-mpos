final _thousands = RegExp(r'\B(?=(\d{3})+(?!\d))');

/// Indonesian Rupiah, no decimals: 1032.0 -> "Rp 1.032". Ties round up, like the Kotlin app.
String rupiah(num amount) {
  final rounded = (amount + 0.5).floor();
  return 'Rp ${rounded.toString().replaceAll(_thousands, '.')}';
}
