import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../preferences.dart';
import 'bluetooth_printer.dart';
import 'sunmi.dart';

enum PrinterStatus {
  ready('ready'),
  outOfPaper('out of paper'),
  overheated('overheated'),
  error('error'),
  unavailable('no printer connected');

  const PrinterStatus(this.label);
  final String label;
}

/// Every receipt printer sits behind this, so adding hardware means one new class.
abstract interface class PrinterPort {
  String get name;

  /// Checked before every print (paper, temperature, connection).
  Future<PrinterStatus> status();

  /// Prints pre-formatted 58 mm lines. Returns [PrinterStatus.ready] on success.
  Future<PrinterStatus> print(List<String> lines);
}

/// Development printer: writes the receipt to the log.
class FakePrinter implements PrinterPort {
  const FakePrinter();

  @override
  String get name => 'Log (development)';

  @override
  Future<PrinterStatus> status() async => PrinterStatus.ready;

  @override
  Future<PrinterStatus> print(List<String> lines) async {
    debugPrint('RECEIPT\n${lines.join('\n')}');
    return PrinterStatus.ready;
  }
}

class NoPrinter implements PrinterPort {
  const NoPrinter();

  @override
  String get name => 'None';

  @override
  Future<PrinterStatus> status() async => PrinterStatus.unavailable;

  @override
  Future<PrinterStatus> print(List<String> lines) async => PrinterStatus.unavailable;
}

/// The paired Bluetooth printer chosen in printer settings, if any.
final savedPrinterProvider = NotifierProvider<SavedPrinter, ({String id, String name})?>(SavedPrinter.new);

class SavedPrinter extends Notifier<({String id, String name})?> {
  @override
  ({String id, String name})? build() {
    final prefs = ref.watch(preferencesProvider);
    final id = prefs.getString('printer_id');
    return id == null ? null : (id: id, name: prefs.getString('printer_name') ?? id);
  }

  void choose(String id, String name) {
    ref.read(preferencesProvider)
      ..setString('printer_id', id)
      ..setString('printer_name', name);
    state = (id: id, name: name);
  }

  void forget() {
    ref.read(preferencesProvider)
      ..remove('printer_id')
      ..remove('printer_name');
    state = null;
  }
}

/// Sunmi built-in printer when present, else the saved Bluetooth printer, else none (log in debug).
final printerProvider = FutureProvider<PrinterPort>((ref) async {
  if (await Sunmi.isSunmi()) return const SunmiPrinter();
  final saved = ref.watch(savedPrinterProvider);
  if (saved != null) return BluetoothPrinter(saved.id, saved.name);
  return kDebugMode ? const FakePrinter() : const NoPrinter();
});
