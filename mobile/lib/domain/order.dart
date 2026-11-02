/// PENDING = confirmed on the device (final number, receipt printed) but not uploaded yet.
enum OrderState { pending, confirmed, failed, cancelled }

typedef OrderLine = ({String productId, String? variantId, int qty});

class OrderRecord {
  const OrderRecord({
    required this.key,
    required this.customerName,
    required this.state,
    required this.orderNumberDisplay,
    required this.referenceId,
    required this.total,
    required this.error,
    required this.createdAt,
    required this.canCancel,
    this.customerTier,
    this.customerQr,
    this.lines = const [],
    this.subtotal,
    this.discount,
    this.promotionName,
  });

  final String key;
  final String customerName;
  final OrderState state;
  final String? orderNumberDisplay;
  final String? referenceId;
  final double? total;
  final String? error;
  final DateTime createdAt;

  /// Uploaded and not cancelled. The server still decides (only while the warehouse hasn't started).
  final bool canCancel;

  final String? customerTier;
  final String? customerQr;
  final List<OrderLine> lines;

  /// Saved since database v3; null on older orders.
  final double? subtotal;
  final double? discount;
  final String? promotionName;

  int get itemCount => lines.fold(0, (sum, l) => sum + l.qty);
}
