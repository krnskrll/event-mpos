import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:universal_ble/universal_ble.dart';

import '../../app/theme.dart';
import '../../app/widgets.dart';
import '../../app/status.dart';
import '../../core/printer/bluetooth_printer.dart';
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
      'Bhoomy POS',
      'Printer test',
      '-' * 32,
      'If you can read this,',
      'printing works.',
    ]);
    setState(() => _message = status == PrinterStatus.ready ? 'Test page printed' : 'Printer ${status.label}');
    ref.invalidate(printerStatusProvider);
  }

  @override
  Widget build(BuildContext context) {
    final printer = ref.watch(printerProvider).value;
    final status = ref.watch(printerStatusProvider).value;
    final saved = ref.watch(savedPrinterProvider);
    final scanning = _scan != null;
    final ready = status == PrinterStatus.ready;
    final kind = switch (printer) {
      SunmiPrinter() => 'Built-in · 58 mm',
      BluetoothPrinter() => 'Bluetooth · 58 mm',
      FakePrinter() => 'Development log (no paper)',
      _ => 'No printer chosen yet',
    };

    return Scaffold(
      body: SafeArea(
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 560),
            child: ListView(
              padding: const EdgeInsets.fromLTRB(8, 8, 16, 24),
              children: [
                const Row(
                  children: [
                    BackButton(),
                    SizedBox(width: 4),
                    Text('Printer', style: AppText.title),
                  ],
                ),
                const SizedBox(height: 12),
                Padding(
                  padding: const EdgeInsets.only(left: 8),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: [
                      _Card(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.stretch,
                          children: [
                            Row(
                              children: [
                                Container(
                                  width: 56,
                                  height: 56,
                                  decoration: BoxDecoration(
                                    color: Brand.greenSoft,
                                    borderRadius: BorderRadius.circular(16),
                                  ),
                                  child: Badge(
                                    smallSize: 10,
                                    alignment: const Alignment(0.55, -0.55),
                                    backgroundColor: ready ? Brand.success : Brand.warning,
                                    child: const Center(
                                      child: Icon(Icons.print_outlined, color: Brand.forest, size: 28),
                                    ),
                                  ),
                                ),
                                const SizedBox(width: 14),
                                Expanded(
                                  child: Column(
                                    crossAxisAlignment: CrossAxisAlignment.start,
                                    children: [
                                      const Text('In use', style: AppText.caption),
                                      Text(
                                        printer?.name ?? 'Checking…',
                                        style: AppText.subhead.copyWith(fontWeight: FontWeight.w900),
                                      ),
                                      Text(kind, style: AppText.caption.copyWith(color: Brand.slate)),
                                    ],
                                  ),
                                ),
                                if (status != null)
                                  ready
                                      ? const Pill('Ready', color: Brand.successInk, background: Brand.successSoft)
                                      : Pill(status.label, color: Brand.warningInk, background: Brand.warningSoft),
                              ],
                            ),
                            if (_message != null) ...[
                              const SizedBox(height: 10),
                              Text(_message!, style: AppText.body.copyWith(color: Brand.slate)),
                            ],
                            const SizedBox(height: 14),
                            Row(
                              children: [
                                Expanded(
                                  child: FilledButton(
                                    onPressed: printer == null ? null : () => _testPrint(printer),
                                    style: FilledButton.styleFrom(minimumSize: const Size(0, 50)),
                                    child: const Text('Print a test page'),
                                  ),
                                ),
                                if (saved != null && printer is! SunmiPrinter) ...[
                                  const SizedBox(width: 8),
                                  FilledButton(
                                    onPressed: ref.read(savedPrinterProvider.notifier).forget,
                                    style: FilledButton.styleFrom(
                                      minimumSize: const Size(0, 50),
                                      backgroundColor: Brand.errorSoft,
                                      foregroundColor: Brand.errorInk,
                                    ),
                                    child: const Text('Forget'),
                                  ),
                                ],
                              ],
                            ),
                          ],
                        ),
                      ),
                      if (printer is! SunmiPrinter) ...[
                        const SizedBox(height: 16),
                        _Card(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.stretch,
                            children: [
                              Row(
                                children: [
                                  Expanded(
                                    child: Column(
                                      crossAxisAlignment: CrossAxisAlignment.start,
                                      children: [
                                        Text(
                                          'Bluetooth printers nearby',
                                          style: AppText.bodyLarge.copyWith(
                                            color: Brand.forest,
                                            fontWeight: FontWeight.w900,
                                          ),
                                        ),
                                        Text(
                                          scanning
                                              ? 'Searching… keep the printer switched on'
                                              : 'Search to find printers around you',
                                          style: AppText.caption,
                                        ),
                                      ],
                                    ),
                                  ),
                                  scanning
                                      ? const SizedBox.square(
                                          dimension: 22,
                                          child: CircularProgressIndicator(strokeWidth: 3),
                                        )
                                      : TextButton(onPressed: _startScan, child: const Text('Search')),
                                ],
                              ),
                              for (final d in _found.values) ...[
                                const Divider(),
                                _FoundPrinter(
                                  d,
                                  inUse: d.deviceId == saved?.id,
                                  onUse: () => ref.read(savedPrinterProvider.notifier).choose(d.deviceId, d.name!),
                                ),
                              ],
                            ],
                          ),
                        ),
                      ],
                      const SizedBox(height: 16),
                      Container(
                        padding: const EdgeInsets.all(14),
                        decoration: BoxDecoration(color: Brand.greenSoft, borderRadius: BorderRadius.circular(14)),
                        child: Row(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            const Icon(Icons.info_outline, size: 18, color: Brand.forest),
                            const SizedBox(width: 10),
                            Expanded(
                              child: Text(
                                'Sunmi devices use their built-in printer automatically. Other devices need a 58 mm Bluetooth receipt printer.',
                                style: AppText.caption.copyWith(color: Brand.forest, fontSize: 13),
                              ),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _Card extends StatelessWidget {
  const _Card({required this.child});

  final Widget child;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Brand.surface,
        borderRadius: BorderRadius.circular(22),
        border: Border.all(color: Brand.cardLine),
        boxShadow: AppShadows.card,
      ),
      child: child,
    );
  }
}

/// A printer found by the Bluetooth search: name, address, signal and Use.
class _FoundPrinter extends StatelessWidget {
  const _FoundPrinter(this.device, {required this.inUse, required this.onUse});

  final BleDevice device;
  final bool inUse;
  final VoidCallback onUse;

  @override
  Widget build(BuildContext context) {
    final rssi = device.rssi ?? -100;
    final bars = rssi >= -60
        ? 4
        : rssi >= -70
        ? 3
        : rssi >= -80
        ? 2
        : 1;
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 8),
      child: Row(
        children: [
          Container(
            width: 40,
            height: 40,
            decoration: BoxDecoration(color: Brand.sunken, borderRadius: BorderRadius.circular(12)),
            child: const Icon(Icons.print_outlined, size: 20, color: Brand.slate),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(device.name!, style: AppText.body.copyWith(fontWeight: FontWeight.w700)),
                Text(device.deviceId, style: AppText.caption),
              ],
            ),
          ),
          Semantics(
            label: '$bars of 4 signal bars',
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.end,
              children: [
                for (var i = 0; i < 4; i++)
                  Container(
                    width: 4,
                    height: 6 + i * 3.5,
                    margin: const EdgeInsets.only(left: 2),
                    decoration: BoxDecoration(
                      color: i < bars ? Brand.green : Brand.divider,
                      borderRadius: BorderRadius.circular(1),
                    ),
                  ),
              ],
            ),
          ),
          const SizedBox(width: 12),
          inUse
              ? Text('In use', style: AppText.label.copyWith(color: Brand.successInk))
              : FilledButton.tonal(
                  onPressed: onUse,
                  style: FilledButton.styleFrom(minimumSize: const Size(64, 40)),
                  child: const Text('Use'),
                ),
        ],
      ),
    );
  }
}
