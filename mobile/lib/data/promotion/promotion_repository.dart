import 'package:decimal/decimal.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/db/database.dart';
import '../../domain/promotion.dart';
import '../session/session.dart';

/// Promotions of the selected event, from the local DB (works offline).
final promotionsProvider = StreamProvider.autoDispose((ref) {
  final db = ref.watch(databaseProvider);
  final eventId = ref.watch(sessionEventProvider)!.id;
  final query = db.select(db.promotions)..where((p) => p.eventId.equals(eventId));
  return query.watch().map((rows) => rows.map(_toDomain).toList());
});

Decimal? _decimal(double? value) => value == null ? null : Decimal.parse(value.toString());

Promotion _toDomain(PromotionRow r) => Promotion(
  name: r.name,
  type: PromotionType.values.asNameMap()[r.type.toLowerCase()] ?? PromotionType.percentage,
  percentOff: _decimal(r.percentOff),
  amountOff: _decimal(r.amountOff),
  productIds: r.scope == 'SPECIFIC' ? r.productIds.split(',').where((id) => id.isNotEmpty).toSet() : const {},
  minQty: r.minQty,
  minSubtotal: _decimal(r.minSubtotal),
  startsAt: DateTime.tryParse(r.startsAt ?? ''),
  endsAt: DateTime.tryParse(r.endsAt ?? ''),
  priority: r.priority,
);
