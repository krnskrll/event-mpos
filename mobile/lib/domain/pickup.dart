/// An order on the pickup board (any device at this location sold it).
class PickupOrder {
  const PickupOrder({
    required this.id,
    required this.number,
    required this.customerName,
    required this.cashierName,
    required this.items,
    required this.status,
    required this.total,
    required this.createdAt,
    this.stockConflict = false,
    this.lateSynced = false,
  });

  factory PickupOrder.fromJson(Map<String, dynamic> json) => PickupOrder(
    id: json['id'],
    number: json['orderNumberDisplay'] ?? '',
    customerName: json['customerName'] ?? '',
    cashierName: json['baName'] ?? '',
    items: [for (final i in json['items'] as List? ?? []) (name: i['name'] as String, qty: i['qty'] as int)],
    status: json['status'],
    total: (json['total'] as num? ?? 0).toDouble(),
    createdAt: DateTime.parse(json['createdAt']).toLocal(),
    stockConflict: json['stockConflict'] ?? false,
    lateSynced: json['lateSynced'] ?? false,
  );

  PickupOrder withStatus(String status) => PickupOrder(
    id: id,
    number: number,
    customerName: customerName,
    cashierName: cashierName,
    items: items,
    status: status,
    total: total,
    createdAt: createdAt,
    stockConflict: stockConflict,
    lateSynced: lateSynced,
  );

  final String id;
  final String number;
  final String customerName;
  final String cashierName;
  final List<({String name, int qty})> items;

  /// NEW → PACKING → READY → COLLECTED, or CANCELLED.
  final String status;
  final double total;
  final DateTime createdAt;
  final bool stockConflict;
  final bool lateSynced;

  /// The status the board's main button moves it to.
  String? get next => const {'NEW': 'PACKING', 'PACKING': 'READY', 'READY': 'COLLECTED'}[status];

  /// Nothing has left the warehouse yet.
  bool get canCancel => status == 'NEW' || status == 'PACKING';
}

/// The board's columns: status, title, and the main button's label.
const pickupColumns = [
  ('NEW', 'New', 'Mark packing'),
  ('PACKING', 'Packing', 'Mark ready'),
  ('READY', 'Ready for pickup', 'Collected'),
];
