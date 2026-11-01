import 'package:flutter_test/flutter_test.dart';
import 'package:mpos/core/format/money.dart';

void main() {
  test('formats rupiah with dot thousands and no decimals', () {
    expect(rupiah(0), 'Rp 0');
    expect(rupiah(1032.0), 'Rp 1.032');
    expect(rupiah(1500000), 'Rp 1.500.000');
    expect(rupiah(999.5), 'Rp 1.000');
    expect(rupiah(-1032), 'Rp -1.032');
  });
}
