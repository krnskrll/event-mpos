import 'package:flutter_test/flutter_test.dart';
import 'package:mpos/domain/receipt.dart';

ReceiptData _sample({double discount = 15400, List<ReceiptLine>? items, String location = 'Bazaar Ramadan GI'}) =>
    ReceiptData(
      store: 'Glow Mart Bandung',
      location: location,
      address: 'Jl. Asia Afrika No. 120, Bandung',
      staffName: 'Bella',
      orderNumberDisplay: '#003',
      referenceId: 'ORD-20260630-EV01-003',
      customerName: 'Aisha Rahman',
      customerTier: 'PLATINUM',
      items:
          items ??
          const [
            ReceiptLine('Daily Mineral Sunscreen SPF50', 1, 99000),
            ReceiptLine('Gentle Foaming Cleanser 150ml', 1, 55000),
          ],
      subtotal: 154000,
      discount: discount,
      promotionName: '10% off everything',
      total: 154000 - discount,
      dateText: '30/06/2026 13:24',
    );

void main() {
  test('store on top, framed order number, staff and customer, Indonesian footer; every line fits 58 mm', () {
    final lines = ReceiptFormatter.format(_sample());
    expect(lines.every((l) => l.length <= ReceiptFormatter.width), isTrue);
    expect(lines.first, '       GLOW MART BANDUNG');
    expect(lines, containsAllInOrder(['Bazaar Ramadan GI'.padLeft(24), '=' * 32, '          PESANAN #003', '=' * 32]));
    expect(lines, contains('30/06/2026 13:24    Kasir: Bella'));
    final text = lines.join('\n');
    for (final field in [
      'Jl. Asia Afrika No. 120,',
      'Pelanggan: Aisha Rahman', '(PLATINUM)',
      'Diskon (10% off',
      'Bayar di kasir',
      'Terima kasih',
      'Bukan faktur pajak',
    ]) {
      expect(text, contains(field));
    }
    expect(lines, contains('TOTAL                 Rp 138.600'));
  });

  test('an outlet named like the store prints the name once', () {
    final lines = ReceiptFormatter.format(_sample(location: 'Glow Mart Bandung'));
    expect(lines.where((l) => l.trim().toUpperCase() == 'GLOW MART BANDUNG'), hasLength(1));
  });

  test('long item names wrap instead of losing the shade', () {
    final lines = ReceiptFormatter.format(
      _sample(items: const [ReceiptLine('Velvet Matte Lipstick — Rose — Plum (limited)', 2, 178000)]),
    );
    expect(lines, containsAll(['2 x Velvet Matte Lipstick — Rose', '— Plum (limited)']));
  });

  test('no discount line without a discount', () {
    expect(ReceiptFormatter.format(_sample(discount: 0)).join('\n'), isNot(contains('Diskon')));
  });
}
