import '../core/format/money.dart';

class ReceiptLine {
  const ReceiptLine(this.name, this.qty, this.lineTotal);
  final String name;
  final int qty;
  final double lineTotal;
}

/// Everything printed on a receipt, captured at confirm time so a reprint needs no network.
class ReceiptData {
  const ReceiptData({
    required this.brand,
    required this.eventName,
    required this.orderNumberDisplay,
    required this.referenceId,
    required this.customerName,
    this.customerTier,
    required this.items,
    required this.subtotal,
    required this.discount,
    this.promotionName,
    required this.total,
    required this.cashierNote,
    required this.dateText,
  });

  final String brand;
  final String eventName;
  final String orderNumberDisplay;
  final String referenceId;
  final String customerName;
  final String? customerTier;
  final List<ReceiptLine> items;
  final double subtotal;
  final double discount;
  final String? promotionName;
  final double total;
  final String cashierNote;
  final String dateText;
}

/// Renders a receipt into 58 mm thermal lines (32 characters), whatever the printer.
abstract final class ReceiptFormatter {
  static const width = 32;

  /// First line of a cancelled order's stored receipt (also printed on any reprint).
  static const cancelled = '       *** CANCELLED ***';

  static List<String> format(ReceiptData r) {
    final lines = [
      _center(r.brand.toUpperCase()),
      _center(r.eventName),
      _divider,
      _leftRight('Order ${r.orderNumberDisplay}', r.dateText),
      r.referenceId,
      ..._wrap('Customer: ${r.customerName}${r.customerTier == null ? '' : ' (${r.customerTier})'}'),
      _divider,
      // Name wrapped on its own lines, amount right-aligned below: shades never get cut off on 58 mm.
      for (final item in r.items) ...[..._wrap('${item.qty} x ${item.name}'), _leftRight('', rupiah(item.lineTotal))],
      _divider,
      _leftRight('Subtotal', rupiah(r.subtotal)),
      if (r.discount > 0)
        _leftRight('Discount${r.promotionName == null ? '' : ' (${r.promotionName})'}', '-${rupiah(r.discount)}'),
      _leftRight('TOTAL', rupiah(r.total)),
      _divider,
      ..._wrap(r.cashierNote).map(_center),
      '',
      _center('Payment at cashier counter'),
      _center('Not a tax invoice'),
    ];
    return [for (final line in lines) line.length > width ? line.substring(0, width) : line];
  }

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
