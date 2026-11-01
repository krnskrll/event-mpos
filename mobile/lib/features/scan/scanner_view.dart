import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:mobile_scanner/mobile_scanner.dart';

import '../../app/theme.dart';

/// Full-screen camera scanner with a target frame and a typed-entry fallback. While [onCode] runs
/// (e.g. a result sheet is open) further reads are ignored.
class ScannerView extends StatefulWidget {
  const ScannerView({
    super.key,
    required this.title,
    required this.hint,
    required this.manualLabel,
    required this.onCode,
    this.formats = const [],
    this.actions = const [],
  });

  final String title;
  final String hint;
  final String manualLabel;
  final Future<void> Function(String code) onCode;
  final List<BarcodeFormat> formats;
  final List<Widget> actions;

  @override
  State<ScannerView> createState() => _ScannerViewState();
}

class _ScannerViewState extends State<ScannerView> {
  late final _controller = MobileScannerController(formats: widget.formats);
  bool _busy = false;

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  Future<void> _handle(String? code) async {
    if (_busy || code == null || code.trim().isEmpty) return;
    _busy = true;
    await widget.onCode(code.trim());
    _busy = false;
  }

  Future<void> _typeCode() async {
    final text = TextEditingController();
    final code = await showDialog<String>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(widget.manualLabel),
        content: TextField(controller: text, autofocus: true, onSubmitted: (v) => Navigator.pop(context, v)),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context), child: const Text('Cancel')),
          TextButton(onPressed: () => Navigator.pop(context, text.text), child: const Text('Search')),
        ],
      ),
    );
    await _handle(code);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.black,
      body: Stack(
        fit: StackFit.expand,
        children: [
          MobileScanner(
            controller: _controller,
            onDetect: (capture) => _handle(capture.barcodes.firstOrNull?.rawValue),
            errorBuilder: (_, _) => const ColoredBox(color: Color(0xFF0A0A0A)),
          ),
          Center(
            child: Container(
              width: 240,
              height: 240,
              decoration: BoxDecoration(
                border: Border.all(color: Brand.teal, width: 3),
                borderRadius: BorderRadius.circular(16),
              ),
            ),
          ),
          SafeArea(
            child: Column(
              children: [
                Row(
                  children: [
                    IconButton(
                      onPressed: context.pop,
                      icon: const Icon(Icons.close, color: Colors.white),
                    ),
                    Expanded(
                      child: Text(
                        widget.title,
                        style: const TextStyle(color: Colors.white, fontSize: 17, fontWeight: FontWeight.bold),
                      ),
                    ),
                    ...widget.actions,
                  ],
                ),
                const Spacer(),
                Padding(
                  padding: const EdgeInsets.all(24),
                  child: Column(
                    children: [
                      Text(
                        widget.hint,
                        textAlign: TextAlign.center,
                        style: const TextStyle(color: Colors.white70, fontSize: 14),
                      ),
                      TextButton.icon(
                        onPressed: _typeCode,
                        icon: const Icon(Icons.keyboard_outlined, size: 18),
                        label: Text(widget.manualLabel),
                        style: TextButton.styleFrom(foregroundColor: const Color(0xFF5BB8F5)),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
