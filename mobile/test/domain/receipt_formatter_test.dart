import 'package:flutter_test/flutter_test.dart';
import 'package:mpos/domain/receipt.dart';

ReceiptData _sample({double discount = 15400, List<ReceiptLine>? items}) => ReceiptData(
  brand: 'Lumière Beauty',
  eventName: 'Mid-Year Beauty Roadshow',
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
  cashierNote: 'Please proceed to the cashier to complete payment.',
  dateText: '30/06 13:24',
);

void main() {
  test('every line fits 58 mm and carries the required fields', () {
    final lines = ReceiptFormatter.format(_sample());
    expect(lines.every((l) => l.length <= ReceiptFormatter.width), isTrue);
    final text = lines.join('\n');
    for (final field in ['LUMIÈRE BEAUTY', 'Mid-Year Beauty Roadshow', '#003', 'Aisha Rahman', 'Discount', 'cashier']) {
      expect(text, contains(field));
    }
    expect(lines, contains('TOTAL                 Rp 138.600'));
  });

  test('long item names wrap instead of losing the shade', () {
    final lines = ReceiptFormatter.format(
      _sample(items: const [ReceiptLine('Velvet Matte Lipstick — Rose — Plum (limited)', 2, 178000)]),
    );
    expect(lines, containsAll(['2 x Velvet Matte Lipstick — Rose', '— Plum (limited)']));
  });

  test('no discount line without a discount', () {
    expect(ReceiptFormatter.format(_sample(discount: 0)).join('\n'), isNot(contains('Discount')));
  });
}
