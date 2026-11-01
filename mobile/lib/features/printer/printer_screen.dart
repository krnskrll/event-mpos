import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:universal_ble/universal_ble.dart';

import '../../app/theme.dart';
import '../../app/widgets.dart';
import '../../core/printer/printer.dart';
import '../../core/printer/sunmi.dart';

/// Shows the active printer, prints a test page, and pairs a Bluetooth printer on non-Sunmi devices.
class PrinterScreen extends ConsumerStatefulWidget {
  const PrinterScreen({super.key});

  @override
  ConsumerState<PrinterScreen> createState() => _PrinterScreenState();
}

class _PrinterScreenState extends ConsumerState<PrinterScreen> {
  final _found = <String, BleDevice>{};
  StreamSubscription<BleDevice>? _scan;
  String? _message;

  @override
  void dispose() {
    _stopScan();
    super.dispose();
  }

  Future<void> _startScan() async {
    setState(_found.clear);
    try {
      await UniversalBle.requestPermissions();
      _scan = UniversalBle.scanStream.listen((d) {
        if (d.name?.isNotEmpty ?? false) setState(() => _found[d.deviceId] = d);
      });
      await UniversalBle.startScan();
      Timer(const Duration(seconds: 8), _stopScan);
    } catch (_) {
      _stopScan();
      setState(() => _message = 'Bluetooth is off or not allowed on this device.');
    }
  }

  void _stopScan() {
    _scan?.cancel();
    UniversalBle.stopScan().catchError((_) {});
    if (mounted) setState(() => _scan = null);
  }

  Future<void> _testPrint(PrinterPort printer) async {
    setState(() => _message = 'Printing…');
    final status = await printer.print([
      'event-mpos',
      'Printer test',
      '-' * 32,
      'If you can read this,',
      'printing works.',
    ]);
    setState(() => _message = status == PrinterStatus.ready ? 'Printed ✓' : 'Printer ${status.label}');
  }

  @override
  Widget build(BuildContext context) {
    final printer = ref.watch(printerProvider).value;
    final saved = ref.watch(savedPrinterProvider);
    final scanning = _scan != null;

    return Scaffold(
      appBar: AppBar(title: const Text('Printer')),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          SectionCard(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text('In use', style: TextStyle(color: Brand.textFaint, fontSize: 12)),
                Text(
                  printer?.name ?? 'Checking…',
                  style: const TextStyle(color: Brand.heading, fontSize: 18, fontWeight: FontWeight.bold),
                ),
                if (_message != null) Text(_message!, style: const TextStyle(color: Brand.textMuted)),
                const SizedBox(height: 12),
                BrandButton(text: 'Print test page', onPressed: printer == null ? null : () => _testPrint(printer)),
              ],
            ),
          ),
          if (printer is! SunmiPrinter) ...[
            const SizedBox(height: 20),
            Row(
              children: [
                const Expanded(
                  child: Text('Bluetooth printers', style: TextStyle(fontWeight: FontWeight.bold)),
                ),
                if (saved != null)
                  TextButton(
                    onPressed: ref.read(savedPrinterProvider.notifier).forget,
                    child: const Text('Forget', style: TextStyle(color: Brand.danger)),
                  ),
                TextButton(onPressed: scanning ? null : _startScan, child: Text(scanning ? 'Scanning…' : 'Scan')),
              ],
            ),
            const Text(
              'Use a 58 mm ESC/POS printer with Bluetooth Low Energy.',
              style: TextStyle(color: Brand.textMuted, fontSize: 12),
            ),
            for (final d in _found.values)
              ListTile(
                title: Text(d.name!),
                subtitle: Text(d.deviceId),
                trailing: d.deviceId == saved?.id ? const Icon(Icons.check, color: Brand.teal) : null,
                onTap: () => ref.read(savedPrinterProvider.notifier).choose(d.deviceId, d.name!),
              ),
          ],
        ],
      ),
    );
  }
}
