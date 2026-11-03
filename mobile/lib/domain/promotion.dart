import 'package:decimal/decimal.dart';

enum PromotionType { percentage, fixed, bundle }

class Promotion {
  const Promotion({
    required this.name,
    required this.type,
    this.percentOff,
    this.amountOff,
    this.productIds = const {},
    this.minQty,
    this.minSubtotal,
    this.startsAt,
    this.endsAt,
    this.priority = 0,
  });

  final String name;
  final PromotionType type;
  final Decimal? percentOff;
  final Decimal? amountOff;

  /// Empty = applies to the whole cart (scope ALL).
  final Set<String> productIds;
  final int? minQty;
  final Decimal? minSubtotal;
  final DateTime? startsAt;
  final DateTime? endsAt;

  /// Tie-breaker between equal discounts (higher wins), as on the server.
  final int priority;
}

class PricedLine {
  const PricedLine(this.productId, this.unitPrice, this.qty);
  final String productId;
  final Decimal unitPrice;
  final int qty;
}

typedef AppliedPromotion = ({Promotion promotion, Decimal discount});

/// Client mirror of backend-go/internal/promotion/engine.go: powers the offline cart preview, the
/// server stays authoritative. Both engines must pass testdata/promotion-cases.json.
abstract final class PromotionEngine {
  /// Largest discount wins; ties go to the higher priority, then the later name.
  static AppliedPromotion? best(List<PricedLine> lines, List<Promotion> promotions, DateTime now) {
    AppliedPromotion? best;
    for (final promotion in promotions) {
      final discount = discountFor(promotion, lines, now);
      if (discount == null || discount <= Decimal.zero) continue;
      final candidate = (promotion: promotion, discount: discount);
      if (best == null || _compare(candidate, best) > 0) best = candidate;
    }
    return best;
  }

  static int _compare(AppliedPromotion a, AppliedPromotion b) => [
    a.discount.compareTo(b.discount),
    a.promotion.priority.compareTo(b.promotion.priority),
    a.promotion.name.compareTo(b.promotion.name),
  ].firstWhere((c) => c != 0, orElse: () => 0);

  static Decimal? discountFor(Promotion p, List<PricedLine> lines, DateTime now) {
    if (p.startsAt != null && now.isBefore(p.startsAt!)) return null;
    if (p.endsAt != null && now.isAfter(p.endsAt!)) return null;

    final applicable = p.productIds.isEmpty ? lines : lines.where((l) => p.productIds.contains(l.productId)).toList();
    if (applicable.isEmpty) return null;
    final qty = applicable.fold(0, (sum, l) => sum + l.qty);
    final subtotal = applicable.fold(Decimal.zero, (sum, l) => sum + l.unitPrice * Decimal.fromInt(l.qty));
    if (p.minQty != null && qty < p.minQty! && p.type != PromotionType.bundle) return null;
    if (p.minSubtotal != null && subtotal < p.minSubtotal!) return null;

    final raw = switch (p.type) {
      PromotionType.percentage =>
        p.percentOff == null ? null : (subtotal * p.percentOff! / Decimal.fromInt(100)).toDecimal(),
      PromotionType.fixed => p.amountOff,
      PromotionType.bundle => p.minQty == null || qty < p.minQty! ? null : p.amountOff,
    };
    return raw?.clamp(Decimal.zero, subtotal).round();
  }
}
