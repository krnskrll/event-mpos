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
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(20),
          child: Column(
            children: [
              CircleAvatar(
                radius: 32,
                backgroundColor: cancelled ? Brand.sunken : Brand.greenSoft,
                child: Icon(
                  cancelled ? Icons.block : Icons.check,
                  color: cancelled ? Brand.slate : Brand.green,
                  size: 34,
                ),
              ),
              const SizedBox(height: 12),
              Text(cancelled ? 'Order cancelled' : 'Order confirmed', style: AppText.heading),
              Text('Receipt printed automatically', style: AppText.caption.copyWith(color: Brand.slate)),
              const SizedBox(height: 16),
              Expanded(
                child: Container(
                  width: double.infinity,
                  padding: const EdgeInsets.all(16),
                  decoration: BoxDecoration(
                    color: Brand.surface,
                    borderRadius: BorderRadius.circular(16),
                    border: Border.all(color: Brand.cardLine),
                  ),
                  child: lines == null
                      ? const Center(child: CircularProgressIndicator())
                      : SingleChildScrollView(
                          child: Text(
                            lines.join('\n'),
                            style: const TextStyle(fontFamily: 'monospace', fontSize: 12, color: Brand.ink),
                          ),
                        ),
                ),
              ),
              if (_message != null)
                Padding(
                  padding: const EdgeInsets.only(top: 8),
                  child: Text(
                    _message!,
                    style: AppText.label.copyWith(color: _failed ? Brand.errorInk : Brand.successInk),
                  ),
                ),
              const SizedBox(height: 12),
              BrandButton(text: 'New order', onPressed: () => context.go('/catalogue')),
              const SizedBox(height: 8),
              Row(
                children: [
                  Expanded(
                    child: BrandButton(
                      text: _printing ? 'Printing…' : 'Reprint',
                      secondary: true,
                      onPressed: _printing ? null : _reprint,
                    ),
                  ),
                  const SizedBox(width: 8),
                  Expanded(
                    child: BrandButton(text: 'View orders', secondary: true, onPressed: () => context.go('/orders')),
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
