import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../app/theme.dart';
import '../../app/widgets.dart';
import '../../data/receipt/receipt_service.dart';
import '../../domain/receipt.dart';

final _receiptProvider = FutureProvider.autoDispose.family(
  (ref, String orderKey) => ref.watch(receiptServiceProvider).lines(orderKey),
);

/// The stored receipt of an order, with reprint (no network needed).
class ReceiptScreen extends ConsumerStatefulWidget {
  const ReceiptScreen(this.orderKey, {super.key});

  final String orderKey;

  @override
  ConsumerState<ReceiptScreen> createState() => _ReceiptScreenState();
}

class _ReceiptScreenState extends ConsumerState<ReceiptScreen> {
  bool _printing = false;
  String? _message;
  bool _failed = false;

  Future<void> _reprint() async {
    setState(() => _printing = true);
    final error = await ref.read(receiptServiceProvider).reprint(widget.orderKey);
    if (!mounted) return;
    setState(() {
      _printing = false;
      _failed = error != null;
      _message = error ?? 'Printed ✓';
    });
  }

  @override
  Widget build(BuildContext context) {
    final lines = ref.watch(_receiptProvider(widget.orderKey)).value;
    final cancelled = lines?.firstOrNull == ReceiptFormatter.cancelled;

    return Scaffold(
      backgroundColor: Brand.bg,
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(20),
          child: Column(
            children: [
              const CircleAvatar(
                radius: 32,
                backgroundColor: Brand.tealSoft,
                child: Icon(Icons.check, color: Brand.teal, size: 34),
              ),
              const SizedBox(height: 12),
              Text(
                cancelled ? 'Order cancelled' : 'Order confirmed',
                style: const TextStyle(color: Brand.heading, fontSize: 20, fontWeight: FontWeight.bold),
              ),
              const Text('Receipt printed automatically', style: TextStyle(color: Brand.textMuted, fontSize: 13)),
              const SizedBox(height: 16),
              Expanded(
                child: Container(
                  width: double.infinity,
                  padding: const EdgeInsets.all(16),
                  decoration: BoxDecoration(color: Colors.white, borderRadius: BorderRadius.circular(12)),
                  child: lines == null
                      ? const Center(child: CircularProgressIndicator(color: Brand.navy))
                      : SingleChildScrollView(
                          child: Text(
                            lines.join('\n'),
                            style: const TextStyle(fontFamily: 'monospace', fontSize: 12, color: Brand.text),
                          ),
                        ),
                ),
              ),
              if (_message != null)
                Padding(
                  padding: const EdgeInsets.only(top: 8),
                  child: Text(_message!, style: TextStyle(color: _failed ? Brand.danger : Brand.teal, fontSize: 13)),
                ),
              const SizedBox(height: 12),
              BrandButton(text: _printing ? 'Printing…' : 'Reprint receipt', loading: _printing, onPressed: _reprint),
              const SizedBox(height: 8),
              Row(
                children: [
                  Expanded(
                    child: BrandButton(text: 'New order', color: Brand.teal, onPressed: () => context.go('/catalogue')),
                  ),
                  const SizedBox(width: 8),
                  Expanded(
                    child: BrandButton(
                      text: 'View orders',
                      color: Brand.textMuted,
                      onPressed: () => context.go('/orders'),
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}
