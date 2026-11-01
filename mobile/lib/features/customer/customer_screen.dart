import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:mobile_scanner/mobile_scanner.dart';

import '../../app/theme.dart';
import '../../app/widgets.dart';
import '../../data/cart/cart_notifier.dart';
import '../../data/member/member_repository.dart';
import '../../domain/customer.dart';
import '../scan/scanner_view.dart';

/// Identifies the customer: member QR (offline cache, then online), or just a typed name.
class CustomerScreen extends ConsumerWidget {
  const CustomerScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    void link(Customer customer) {
      ref.read(cartProvider.notifier).setCustomer(customer);
      context.pop();
    }

    Future<void> identify(String code) async {
      final customer = await ref.read(memberRepositoryProvider).identify(code);
      if (!context.mounted) return;
      final confirmed = customer == null
          ? await _askName(context, code)
          : await showModalBottomSheet<Customer>(context: context, builder: (_) => _ProfileSheet(customer));
      if (confirmed != null) link(confirmed);
    }

    return ScannerView(
      title: 'Scan Member QR',
      hint: "Point the camera at the member's QR code",
      manualLabel: 'Enter member ID / name',
      formats: const [BarcodeFormat.qrCode],
      onCode: identify,
      actions: [
        TextButton(
          onPressed: context.pop,
          child: const Text('Skip', style: TextStyle(color: Colors.white70)),
        ),
      ],
    );
  }

  /// Unknown QR or ID: continue with just a name.
  static Future<Customer?> _askName(BuildContext context, String code) async {
    final name = TextEditingController(text: code.startsWith('MBR-') ? '' : code);
    final typed = await showDialog<String>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text('No member for "$code"'),
        content: TextField(
          controller: name,
          autofocus: true,
          decoration: const InputDecoration(hintText: 'Customer name'),
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context), child: const Text('Scan again')),
          TextButton(onPressed: () => Navigator.pop(context, name.text.trim()), child: const Text('Use this name')),
        ],
      ),
    );
    return typed == null || typed.isEmpty ? null : Customer(name: typed, source: CustomerSource.manual);
  }
}

class _ProfileSheet extends StatelessWidget {
  const _ProfileSheet(this.customer);

  final Customer customer;

  @override
  Widget build(BuildContext context) {
    final source = customer.source == CustomerSource.localCache ? 'matched offline' : 'matched online';
    return SafeArea(
      child: Padding(
        padding: const EdgeInsets.fromLTRB(24, 24, 24, 16),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const CircleAvatar(
              radius: 32,
              backgroundColor: Brand.tealSoft,
              child: Icon(Icons.check_circle, color: Brand.teal, size: 36),
            ),
            const SizedBox(height: 14),
            Text(
              customer.name,
              style: const TextStyle(color: Brand.heading, fontSize: 20, fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 6),
            Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                if (customer.tier != null) ...[
                  Pill(customer.tier!, color: Brand.gold, background: Brand.goldBg),
                  const SizedBox(width: 8),
                ],
                Text(source, style: const TextStyle(color: Brand.textFaint, fontSize: 12)),
              ],
            ),
            if (customer.qrCode != null)
              Text(customer.qrCode!, style: const TextStyle(color: Brand.textMuted, fontSize: 12)),
            const SizedBox(height: 20),
            BrandButton(text: 'Confirm & Continue', onPressed: () => Navigator.pop(context, customer)),
            TextButton(
              onPressed: () => Navigator.pop(context),
              child: const Text('Scan a different member', style: TextStyle(color: Brand.teal)),
            ),
          ],
        ),
      ),
    );
  }
}
