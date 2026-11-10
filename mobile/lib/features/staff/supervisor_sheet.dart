import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../app/theme.dart';
import '../../app/widgets.dart';
import '../../data/device/device_repository.dart';
import '../../data/session/session.dart';
import '../../domain/staff.dart';
import 'pin_pad.dart';

const cancelReasons = ['Wrong entry', 'Customer cancelled', 'Out of stock'];

/// Sends the approval to the server; returns an error to show, or null when it went through.
typedef Approve = Future<String?> Function(String reason, StaffMember supervisor, String pin);

/// A supervisor of this location approves a cancel with their PIN (always typed, even when they are
/// the one signed in). A sheet on phones, a dialog on tablets. Returns the approving supervisor.
Future<StaffMember?> showSupervisorSheet(BuildContext context, {required String subtitle, required Approve approve}) {
  final content = SupervisorApproval(subtitle: subtitle, approve: approve);
  if (isWide(context)) {
    return showDialog<StaffMember>(
      context: context,
      builder: (_) => Dialog(
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 440),
          child: SingleChildScrollView(padding: const EdgeInsets.all(20), child: content),
        ),
      ),
    );
  }
  return showModalBottomSheet<StaffMember>(
    context: context,
    useRootNavigator: true,
    isScrollControlled: true,
    builder: (_) => SafeArea(
      child: SingleChildScrollView(padding: const EdgeInsets.fromLTRB(20, 0, 20, 20), child: content),
    ),
  );
}

class SupervisorApproval extends ConsumerStatefulWidget {
  const SupervisorApproval({super.key, required this.subtitle, required this.approve});

  final String subtitle;
  final Approve approve;

  @override
  ConsumerState<SupervisorApproval> createState() => _SupervisorApprovalState();
}

class _SupervisorApprovalState extends ConsumerState<SupervisorApproval> {
  String _reason = cancelReasons[1];
  StaffMember? _supervisor;
  String _pin = '';
  String? _error;
  bool _sending = false;

  Future<void> _confirm(StaffMember supervisor) async {
    setState(() {
      _sending = true;
      _error = null;
    });
    final error = await widget.approve(_reason, supervisor, _pin);
    if (!mounted) return;
    if (error == null) return Navigator.pop(context, supervisor);
    setState(() {
      _sending = false;
      _pin = '';
      _error = error;
    });
  }

  @override
  Widget build(BuildContext context) {
    final supervisors = ref.watch(supervisorsProvider);
    final location = ref.watch(sessionEventProvider)?.name ?? '';
    final chosen = supervisors.where((s) => s.id == _supervisor?.id).firstOrNull ?? supervisors.firstOrNull;
    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text('Supervisor approval needed', style: AppText.heading.copyWith(fontWeight: FontWeight.w900)),
                  Text(widget.subtitle, style: AppText.caption.copyWith(fontSize: 13)),
                ],
              ),
            ),
            IconButton(
              onPressed: () => Navigator.pop(context),
              tooltip: 'Close',
              style: IconButton.styleFrom(backgroundColor: Brand.sunken),
              icon: const Icon(Icons.close),
            ),
          ],
        ),
        const SizedBox(height: 12),
        if (chosen == null) ...[
          Container(
            padding: const EdgeInsets.all(14),
            decoration: BoxDecoration(color: Brand.goldSoft, borderRadius: BorderRadius.circular(14)),
            child: Text(
              "There's no supervisor at this location yet. Ask the store owner to add one in Bhoomy.",
              style: AppText.body.copyWith(color: Brand.goldInk),
            ),
          ),
          const SizedBox(height: 16),
          BrandButton(text: 'Close', secondary: true, onPressed: () => Navigator.pop(context)),
        ] else ...[
          DropdownButtonFormField<String>(
            initialValue: _reason,
            decoration: const InputDecoration(labelText: 'Reason'),
            items: [for (final r in cancelReasons) DropdownMenuItem(value: r, child: Text(r))],
            onChanged: (r) => setState(() => _reason = r ?? _reason),
          ),
          const SizedBox(height: 14),
          Text('Supervisors at $location', style: AppText.label),
          const SizedBox(height: 6),
          GridView(
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
              crossAxisCount: 2,
              mainAxisSpacing: 8,
              crossAxisSpacing: 8,
              mainAxisExtent: 56,
            ),
            children: [for (final s in supervisors) _choice(s, s.id == chosen.id)],
          ),
          const SizedBox(height: 14),
          PinDots(_pin.length, error: _error != null),
          const SizedBox(height: 6),
          Text(
            _error ?? 'PIN of the selected supervisor',
            textAlign: TextAlign.center,
            style: AppText.label.copyWith(color: _error != null ? Brand.errorStrong : Brand.faint),
          ),
          const SizedBox(height: 6),
          NumberPad(
            keyColor: Brand.sunken,
            onDigit: _sending
                ? null
                : (d) => setState(() {
                    if (_pin.length < 6) _pin += d;
                    _error = null;
                  }),
            onBackspace: () => setState(() => _pin = _pin.isEmpty ? '' : _pin.substring(0, _pin.length - 1)),
          ),
          const SizedBox(height: 10),
          SizedBox(
            height: 56,
            child: FilledButton(
              onPressed: _pin.length >= 4 && !_sending ? () => _confirm(chosen) : null,
              style: FilledButton.styleFrom(
                backgroundColor: Brand.errorStrong,
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
                textStyle: AppText.button.copyWith(fontSize: 16),
              ),
              child: _sending
                  ? const SizedBox.square(
                      dimension: 20,
                      child: CircularProgressIndicator(color: Colors.white, strokeWidth: 2),
                    )
                  : const Text('Approve and cancel'),
            ),
          ),
        ],
      ],
    );
  }

  Widget _choice(StaffMember s, bool selected) => Semantics(
    selected: selected,
    button: true,
    child: Material(
      color: selected ? Brand.greenSoft : Brand.surface,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(14),
        side: BorderSide(color: selected ? Brand.green : Brand.divider, width: 2),
      ),
      child: InkWell(
        borderRadius: BorderRadius.circular(14),
        onTap: () => setState(() {
          _supervisor = s;
          _pin = '';
          _error = null;
        }),
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 10),
          child: Row(
            children: [
              StaffAvatar(s, radius: 16),
              const SizedBox(width: 8),
              Expanded(
                child: Text(s.name, maxLines: 1, overflow: TextOverflow.ellipsis, style: AppText.label),
              ),
            ],
          ),
        ),
      ),
    ),
  );
}
