import 'event.dart';
import 'receipt.dart';

/// This device as paired in Bhoomy: its name, its store and the one location it sells at.
/// Same shape as the server's pairing reply and `GET /api/device`.
class PairedDevice {
  const PairedDevice({
    required this.id,
    required this.name,
    required this.storeName,
    required this.location,
    this.removesAt,
    this.receipt = const ReceiptSettings(),
  });

  factory PairedDevice.fromJson(Map<String, dynamic> json) => PairedDevice(
    id: json['device']['id'],
    name: json['device']['name'],
    removesAt: DateTime.tryParse(json['device']['removesAt'] ?? ''),
    storeName: json['storeName'] ?? '',
    location: AssignedEvent.fromJson(json['location']),
    receipt: ReceiptSettings.fromJson(json['receipt']),
  );

  Map<String, dynamic> toJson() => {
    'device': {'id': id, 'name': name, 'removesAt': ?removesAt?.toIso8601String()},
    'storeName': storeName,
    'location': location.toJson(),
    'receipt': receipt.toJson(),
  };

  final String id;
  final String name;
  final String storeName;
  final AssignedEvent location;

  /// Set once the event closed: the device stops working then.
  final DateTime? removesAt;

  /// The store's Pengaturan Struk.
  final ReceiptSettings receipt;
}
