import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../app/theme.dart';
import '../../data/receipt/receipt_service.dart';

/// The stored 58 mm receipt lines of an order.
final receiptLinesProvider = FutureProvider.autoDispose.family(
  (ref, String orderKey) => ref.watch(receiptServiceProvider).lines(orderKey),
);

/// Whether the order's receipt printed; null when there is none.
final receiptPrintedProvider = FutureProvider.autoDispose.family(
  (ref, String orderKey) => ref.watch(receiptServiceProvider).printed(orderKey),
);

/// Prints the stored receipt again (no network needed) and says how it went.
Future<void> reprintReceipt(BuildContext context, WidgetRef ref, String orderKey) async {
  final messenger = ScaffoldMessenger.of(context);
  final error = await ref.read(receiptServiceProvider).reprint(orderKey);
  ref.invalidate(receiptPrintedProvider(orderKey));
  messenger.showSnackBar(SnackBar(content: Text(error ?? 'Receipt printed')));
}

/// The receipt as a centred paper slip, aligned exactly as it prints.
class ReceiptPaper extends StatelessWidget {
  const ReceiptPaper(this.lines, {super.key});

  final List<String> lines;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.only(bottom: 8),
        child: Stack(
          clipBehavior: Clip.none,
          children: [
            Container(
              padding: const EdgeInsets.fromLTRB(18, 16, 18, 24),
              decoration: const BoxDecoration(
                color: Brand.surface,
                borderRadius: BorderRadius.vertical(top: Radius.circular(6)),
                boxShadow: AppShadows.paper,
              ),
              child: Text(
                lines.join('\n'),
                style: const TextStyle(
                  fontFamily: 'monospace',
                  fontFamilyFallback: ['Menlo', 'Courier'],
                  fontSize: 11.5,
                  height: 1.5,
                  color: Brand.ink,
                ),
              ),
            ),
            // Torn edge: canvas-coloured circles bite into the bottom.
            Positioned(
              left: 2,
              right: 2,
              bottom: -7,
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [for (var i = 0; i < 12; i++) const CircleAvatar(radius: 7, backgroundColor: Brand.canvas)],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
