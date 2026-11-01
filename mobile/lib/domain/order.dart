/// PENDING = confirmed on the device (final number, receipt printed) but not uploaded yet.
enum OrderState { pending, confirmed, failed, cancelled }

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
}
