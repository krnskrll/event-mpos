/// How the customer was identified, shown so the BA knows whether it was a confirmed match.
enum CustomerSource { localCache, onlineLookup, manual }

/// The customer linked to the current sale.
class Customer {
  const Customer({required this.name, this.tier, this.qrCode, required this.source});

  /// The customer of a sale nobody was identified for (Skip, or no customer added).
  static const walkIn = Customer(name: 'Walk-in', source: CustomerSource.manual);

  factory Customer.fromJson(Map<String, dynamic> json) => Customer(
    name: json['name'],
    tier: json['tier'],
    qrCode: json['qrCode'],
    source: CustomerSource.values.byName(json['source']),
  );

  final String name;
  final String? tier;
  final String? qrCode;
  final CustomerSource source;

  Map<String, dynamic> toJson() => {'name': name, 'tier': tier, 'qrCode': qrCode, 'source': source.name};
}
