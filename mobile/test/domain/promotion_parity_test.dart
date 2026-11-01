import 'dart:convert';
import 'dart:io';

import 'package:decimal/decimal.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mpos/domain/promotion.dart';

/// Runs the fixture shared with backend-go/internal/promotion/parity_test.go, so a change to only
/// one engine fails CI.
void main() {
  final fixture = jsonDecode(File('../testdata/promotion-cases.json').readAsStringSync());
  final now = DateTime.parse(fixture['now']);
  Decimal? dec(Object? v) => v == null ? null : Decimal.parse('$v');
  final carts = {
    for (final MapEntry(:key, :value) in (fixture['carts'] as Map).entries)
      key: [for (final l in value) PricedLine(l['productId'], dec(l['unitPrice'])!, l['qty'])],
  };

  for (final c in fixture['cases']) {
    test(c['name'], () {
      final promotions = [
        for (final p in c['promotions'])
          Promotion(
            name: p['name'],
            type: PromotionType.values.byName((p['type'] as String).toLowerCase()),
            percentOff: dec(p['percentOff']),
            amountOff: dec(p['amountOff']),
            productIds: p['scope'] == 'SPECIFIC' ? Set<String>.from(p['productIds']) : const {},
            minQty: p['minQty'],
            minSubtotal: dec(p['minSubtotal']),
            startsAt: p['startsAt'] == null ? null : DateTime.parse(p['startsAt']),
            endsAt: p['endsAt'] == null ? null : DateTime.parse(p['endsAt']),
            priority: p['priority'] ?? 0,
          ),
      ];
      final best = PromotionEngine.best(carts[c['cart']]!, promotions, now);
      expect(best?.promotion.name, c['expected']['promotion']);
      expect(best?.discount ?? Decimal.zero, dec(c['expected']['discount']));
    });
  }
}
