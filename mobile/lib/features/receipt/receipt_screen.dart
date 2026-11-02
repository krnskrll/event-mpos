import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../app/theme.dart';
import '../../app/widgets.dart';
import '../../core/format/money.dart';
import '../../domain/order.dart';
import '../../domain/receipt.dart';
import '../orders/orders_screen.dart';
import 'receipt_paper.dart';

/// Shown right after Confirm: the order's number, whether the receipt printed, where the upload is,
/// and the receipt itself. "New order" is the next step.
class ReceiptScreen extends ConsumerWidget {
  const ReceiptScreen(this.orderKey, {super.key});

  final String orderKey;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final lines = ref.watch(receiptLinesProvider(orderKey)).value;
    final printed = ref.watch(receiptPrintedProvider(orderKey)).value;
    final order = (ref.watch(ordersProvider).value ?? const <OrderRecord>[])
        .where((o) => o.key == orderKey)
        .firstOrNull;
    final cancelled = lines?.firstOrNull == ReceiptFormatter.cancelled || order?.state == OrderState.cancelled;
    final number = order?.orderNumberDisplay ?? '';
    final summary = [
      printed == false ? 'Receipt not printed' : 'Receipt printed',
      ?order?.customerName,
      if (order?.total != null) rupiah(order!.total!),
    ].join(' · ');
    final (status, ink, soft) = switch (order?.state) {
      OrderState.confirmed => ('Processing', Brand.successInk, Brand.successSoft),
      OrderState.failed => ('Not accepted by the server', Brand.errorInk, Brand.errorSoft),
      OrderState.cancelled => ('Cancelled', Brand.slate, Brand.sunken),
      _ => ('Sending to server', Brand.infoInk, Brand.infoSoft),
    };

    return Scaffold(
      body: SafeArea(
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 480),
            child: Padding(
              padding: const EdgeInsets.fromLTRB(20, 24, 20, 20),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  Center(
                    child: Container(
                      decoration: BoxDecoration(
                        shape: BoxShape.circle,
                        border: Border.all(color: cancelled ? Brand.sunken : Brand.greenSoft, width: 8),
                      ),
                      child: CircleAvatar(
                        radius: 32,
                        backgroundColor: cancelled ? Brand.slate : Brand.green,
                        child: Icon(cancelled ? Icons.block : Icons.check_rounded, color: Colors.white, size: 34),
                      ),
                    ),
                  ),
                  const SizedBox(height: 14),
                  Text(
                    cancelled ? 'Order $number cancelled' : 'Order $number confirmed',
                    textAlign: TextAlign.center,
                    style: AppText.title,
                  ),
                  Text(
                    summary,
                    textAlign: TextAlign.center,
                    style: AppText.caption.copyWith(color: Brand.slate),
                  ),
                  const SizedBox(height: 8),
                  Center(
                    child: Pill(status, color: ink, background: soft),
                  ),
                  const SizedBox(height: 16),
                  Expanded(
                    child: lines == null
                        ? const Center(child: CircularProgressIndicator())
                        : SingleChildScrollView(child: ReceiptPaper(lines)),
                  ),
                  const SizedBox(height: 16),
                  BrandButton(text: 'New order', onPressed: () => context.go('/sell')),
                  const SizedBox(height: 8),
                  Row(
                    children: [
                      Expanded(
                        child: BrandButton(
                          text: 'Reprint',
                          secondary: true,
                          onPressed: () => reprintReceipt(context, ref, orderKey),
                        ),
                      ),
                      const SizedBox(width: 8),
                      Expanded(
                        child: BrandButton(
                          text: 'View orders',
                          secondary: true,
                          onPressed: () => context.go('/orders'),
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}
