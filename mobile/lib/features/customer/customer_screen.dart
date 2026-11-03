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
      final confirmed = await showModalBottomSheet<Customer>(
        context: context,
        isScrollControlled: true, // the name field lifts above the keyboard
        builder: (_) => customer == null ? _UnknownSheet(code) : _MemberSheet(customer),
      );
      if (confirmed != null) link(confirmed);
    }

    return ScannerView(
      title: 'Scan member QR',
      hint: "Point at the member's QR code",
      manualLabel: 'Type member ID or name',
      formats: const [BarcodeFormat.qrCode],
      frame: const Size(220, 220),
      onCode: identify,
      actions: [
        TextButton(
          onPressed: context.pop,
          style: TextButton.styleFrom(
            foregroundColor: Colors.white,
            backgroundColor: Colors.white.withValues(alpha: 0.14),
          ),
          child: const Text('Skip'),
        ),
      ],
    );
  }
}

/// A member matched: who they are and their tier; one tap links them to the order.
class _MemberSheet extends StatelessWidget {
  const _MemberSheet(this.customer);

  final Customer customer;

  @override
  Widget build(BuildContext context) {
    final where = customer.source == CustomerSource.localCache ? 'found on this device' : 'found online';
    return SafeArea(
      child: Padding(
        padding: const EdgeInsets.fromLTRB(22, 8, 22, 16),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Stack(
              clipBehavior: Clip.none,
              children: [
                Container(
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    border: Border.all(color: Brand.greenSoft, width: 8),
                  ),
                  child: CircleAvatar(
                    radius: 38,
                    backgroundColor: Brand.green,
                    child: Text(initials(customer.name), style: AppText.title.copyWith(color: Colors.white)),
                  ),
                ),
                const Positioned(
                  right: 2,
                  bottom: 2,
                  child: CircleAvatar(
                    radius: 14,
                    backgroundColor: Brand.surface,
                    child: Icon(Icons.check_circle, color: Brand.green, size: 26),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 14),
            Text(customer.name, textAlign: TextAlign.center, style: AppText.title.copyWith(fontSize: 22)),
            if (customer.tier != null) ...[
              const SizedBox(height: 6),
              Pill(memberLabel(customer.tier!), color: Brand.goldInk, background: Brand.goldSoft),
            ],
            const SizedBox(height: 6),
            Text([?customer.qrCode, where].join(' · '), style: AppText.caption),
            const SizedBox(height: 20),
            BrandButton(text: 'Use this member', onPressed: () => Navigator.pop(context, customer)),
            TextButton(onPressed: () => Navigator.pop(context), child: const Text('Scan someone else')),
          ],
        ),
      ),
    );
  }
}

/// No member for the code: carry on with just the customer's name.
class _UnknownSheet extends StatefulWidget {
  const _UnknownSheet(this.code);

  final String code;

  @override
  State<_UnknownSheet> createState() => _UnknownSheetState();
}

class _UnknownSheetState extends State<_UnknownSheet> {
  // A typed name (not an ID) is most likely the customer's name already.
  late final _name = TextEditingController(text: widget.code.toUpperCase().startsWith('MBR-') ? '' : widget.code);

  @override
  void dispose() {
    _name.dispose();
    super.dispose();
  }

  void _continue() {
    final name = _name.text.trim();
    if (name.isNotEmpty) Navigator.pop(context, Customer(name: name, source: CustomerSource.manual));
  }

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: EdgeInsets.only(bottom: MediaQuery.viewInsetsOf(context).bottom),
      child: SafeArea(
        child: Padding(
          padding: const EdgeInsets.fromLTRB(22, 8, 22, 16),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Row(
                children: [
                  Container(
                    width: 52,
                    height: 52,
                    decoration: BoxDecoration(color: Brand.goldSoft, borderRadius: BorderRadius.circular(16)),
                    child: const Icon(Icons.person_search_outlined, color: Brand.goldInk, size: 26),
                  ),
                  const SizedBox(width: 14),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'No member for ${widget.code}',
                          style: AppText.subhead.copyWith(fontWeight: FontWeight.w900),
                        ),
                        Text(
                          "Check the ID, or continue with the customer's name.",
                          style: AppText.caption.copyWith(color: Brand.slate),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 16),
              const Text('Customer name', style: AppText.label),
              const SizedBox(height: 6),
              TextField(
                controller: _name,
                autofocus: true,
                textCapitalization: TextCapitalization.words,
                textInputAction: TextInputAction.done,
                onSubmitted: (_) => _continue(),
              ),
              const SizedBox(height: 16),
              BrandButton(text: 'Continue with this name', onPressed: _continue),
              const SizedBox(height: 8),
              BrandButton(text: 'Scan again', secondary: true, onPressed: () => Navigator.pop(context)),
            ],
          ),
        ),
      ),
    );
  }
}
