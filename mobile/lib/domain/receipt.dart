import '../core/format/money.dart';

class ReceiptLine {
  const ReceiptLine(this.name, this.qty, this.lineTotal);
  final String name;
  final int qty;
  final double lineTotal;
}

/// The store's Pengaturan Struk from Bhoomy, kept with the pairing data so receipts print offline.
class ReceiptSettings {
  const ReceiptSettings({
    this.phone = '',
    this.instagram = '',
    this.note = '',
    this.showAddress = true,
    this.showStaff = true,
    this.showCustomer = true,
  });

  /// Missing (device data from before Pengaturan Struk) = the defaults.
  factory ReceiptSettings.fromJson(Map<String, dynamic>? json) => json == null
      ? const ReceiptSettings()
      : ReceiptSettings(
          phone: json['phone'] ?? '',
          instagram: json['instagram'] ?? '',
          note: json['note'] ?? '',
          showAddress: json['showAddress'] ?? true,
          showStaff: json['showStaff'] ?? true,
          showCustomer: json['showCustomer'] ?? true,
        );

  Map<String, dynamic> toJson() => {
    'phone': phone,
    'instagram': instagram,
    'note': note,
    'showAddress': showAddress,
    'showStaff': showStaff,
    'showCustomer': showCustomer,
  };

  final String phone;
  final String instagram;
  final String note;
  final bool showAddress;
  final bool showStaff;
  final bool showCustomer;

  /// Phone and @instagram, on one line when they fit.
  List<String> get contact => [if (phone.isNotEmpty) phone, if (instagram.isNotEmpty) '@$instagram'];
}

/// Everything printed on a receipt, captured at confirm time so a reprint needs no network.
class ReceiptData {
  const ReceiptData({
    required this.store,
    required this.location,
    this.address,
    required this.staffName,
    required this.orderNumberDisplay,
    required this.referenceId,
    required this.customerName,
    this.customerTier,
    required this.items,
    required this.subtotal,
    required this.discount,
    this.promotionName,
    required this.total,
    required this.dateText,
    this.settings = const ReceiptSettings(),
  });

  final String store;
  final String location;
  final String? address;
  final String staffName;
  final String orderNumberDisplay;
  final String referenceId;
  final String customerName;
  final String? customerTier;
  final List<ReceiptLine> items;
  final double subtotal;
  final double discount;
  final String? promotionName;
  final double total;
  final String dateText;
  final ReceiptSettings settings;
}

/// Renders the order slip into 58 mm thermal lines (32 characters), whatever the printer: store and place
/// on top, the order number framed so the cashier finds it, then items, totals and the pay-at-cashier note.
abstract final class ReceiptFormatter {
  static const width = 32;

  /// First line of a cancelled order's stored receipt (also printed on any reprint).
  static const cancelled = '       *** CANCELLED ***';

  static List<String> format(ReceiptData r) {
    final s = r.settings;
    final staff = 'Kasir: ${r.staffName}';
    final contact = s.contact.join(' · ');
    final lines = [
      ..._wrap(r.store.toUpperCase()).map(_center),
      if (r.location.trim().toLowerCase() != r.store.trim().toLowerCase()) ..._wrap(r.location).map(_center),
      if (s.showAddress && (r.address?.trim().isNotEmpty ?? false)) ..._wrap(r.address!.trim()).map(_center),
      if (contact.length <= width && contact.isNotEmpty) _center(contact) else ...s.contact.map(_center),
      _rule,
      _center('PESANAN ${r.orderNumberDisplay}'),
      _rule,
      if (!s.showStaff)
        r.dateText
      else if (r.dateText.length + staff.length < width)
        _leftRight(r.dateText, staff)
      else ...[
        r.dateText,
        ..._wrap(staff),
      ],
      r.referenceId,
      if (s.showCustomer)
        ..._wrap('Pelanggan: ${r.customerName}${r.customerTier == null ? '' : ' (${r.customerTier})'}'),
      _divider,
      // Name wrapped on its own lines, amount right-aligned below: shades never get cut off on 58 mm.
      for (final item in r.items) ...[..._wrap('${item.qty} x ${item.name}'), _leftRight('', rupiah(item.lineTotal))],
      _divider,
      _leftRight('Subtotal', rupiah(r.subtotal)),
      if (r.discount > 0)
        _leftRight('Diskon${r.promotionName == null ? '' : ' (${r.promotionName})'}', '-${rupiah(r.discount)}'),
      _leftRight('TOTAL', rupiah(r.total)),
      _divider,
      _center('Bayar di kasir dengan struk ini'),
      if (s.note.trim().isNotEmpty) ..._wrap(s.note.trim()).map(_center),
      _center('Terima kasih'),
      '',
      _center('Bukan faktur pajak'),
    ];
    return [for (final line in lines) line.length > width ? line.substring(0, width) : line];
  }

  static final _rule = '=' * width;
  static final _divider = '-' * width;

  static String _center(String text) {
    final t = text.length > width ? text.substring(0, width) : text;
    return ' ' * ((width - t.length) ~/ 2) + t;
  }

  static String _leftRight(String left, String right) {
    final room = width - right.length - 1;
    final l = left.length > room ? left.substring(0, room) : left;
    return l + ' ' * (width - l.length - right.length) + right;
  }

  static List<String> _wrap(String text) {
    final lines = <String>[];
    var line = '';
    for (final word in text.split(' ')) {
      if (line.isEmpty) {
        line = word;
      } else if (line.length + 1 + word.length <= width) {
        line += ' $word';
      } else {
        lines.add(line);
        line = word;
      }
    }
    return [...lines, if (line.isNotEmpty) line];
  }
}
