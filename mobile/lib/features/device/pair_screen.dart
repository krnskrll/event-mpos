import 'package:drift/drift.dart' hide Column;
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../app/logo.dart';
import '../../app/theme.dart';
import '../../app/widgets.dart';
import '../../core/db/database.dart';
import '../../core/network/api_failure.dart';
import '../../core/secure/device_storage.dart';
import '../../data/device/device_repository.dart';
import '../../data/session/session.dart';
import '../../domain/event.dart';
import '../staff/pin_pad.dart';

/// Orders confirmed on this device that never reached the server, for the removed-device notice.
final _unsentProvider = FutureProvider.autoDispose((ref) async {
  final db = ref.watch(databaseProvider);
  final q = db.selectOnly(db.pendingOrders)
    ..addColumns([countAll()])
    ..where(db.pendingOrders.status.isIn(['PENDING', 'UPLOADING']));
  return (await q.getSingle()).read(countAll()) ?? 0;
});

/// First run: link the device to a location with the 6-digit code from Bhoomy, download what
/// selling needs, then hand over to the lock screen.
class PairScreen extends ConsumerStatefulWidget {
  const PairScreen({super.key});

  @override
  ConsumerState<PairScreen> createState() => _PairScreenState();
}

class _PairScreenState extends ConsumerState<PairScreen> {
  String _code = '';
  String? _error;
  bool _busy = false;
  bool _paired = false;
  ({int products, int promotions, int staff})? _counts;

  void _type(String digit) => setState(() {
    if (_code.length < 6) _code += digit;
    _error = null;
  });

  void _back() => setState(() {
    if (_code.isNotEmpty) _code = _code.substring(0, _code.length - 1);
    _error = null;
  });

  Future<void> _connect() async {
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      await ref.read(deviceRepositoryProvider).pair(_code);
      if (!mounted) return;
      setState(() => _paired = true);
      await _prepare();
    } on ApiFailure catch (e) {
      if (mounted) {
        setState(() {
          _busy = false;
          _error = e.message;
        });
      }
    }
  }

  Future<void> _prepare() async {
    setState(() {
      _busy = true;
      _error = null;
    });
    try {
      final counts = await ref.read(deviceRepositoryProvider).prepare();
      if (mounted) setState(() => _counts = counts);
    } on ApiFailure catch (e) {
      if (mounted) setState(() => _error = e.message);
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final done = _counts != null;
    return AnnotatedRegion(
      value: SystemUiOverlayStyle.light.copyWith(statusBarColor: Colors.transparent),
      child: Scaffold(
        body: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            _Header(
              title: done ? 'Device ready' : 'Connect this device',
              subtitle: done ? 'This step happens only once.' : 'Only once, when the app is first opened.',
            ),
            Expanded(
              child: Center(
                child: SingleChildScrollView(
                  padding: const EdgeInsets.all(20),
                  child: ConstrainedBox(
                    constraints: const BoxConstraints(maxWidth: 440),
                    child: done
                        ? _Ready(_counts!)
                        : _paired
                        ? _Preparing(error: _error, onRetry: _busy ? null : _prepare)
                        : _entry(),
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _entry() {
    final removed = ref.watch(deviceStorageProvider).removed;
    final unsent = removed ? ref.watch(_unsentProvider).value ?? 0 : 0;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        if (removed) ...[
          Container(
            padding: const EdgeInsets.all(14),
            decoration: BoxDecoration(color: Brand.errorSoft, borderRadius: BorderRadius.circular(14)),
            child: Text(
              [
                'This device was removed in Bhoomy. Connect it again with a new code.',
                if (unsent > 0) '$unsent ${unsent == 1 ? 'order was' : 'orders were'} not sent.',
              ].join(' '),
              style: AppText.body.copyWith(color: Brand.errorInk),
            ),
          ),
          const SizedBox(height: 16),
        ],
        Text(
          'Enter the device code to link this device.',
          textAlign: TextAlign.center,
          style: AppText.body.copyWith(color: Brand.slate),
        ),
        const SizedBox(height: 16),
        Semantics(
          label: 'Device code, ${_code.length} of 6 digits',
          child: Row(
            children: [
              for (var i = 0; i < 6; i++)
                Expanded(
                  child: Container(
                    height: 58,
                    margin: const EdgeInsets.symmetric(horizontal: 4),
                    alignment: Alignment.center,
                    decoration: BoxDecoration(
                      color: Brand.surface,
                      borderRadius: BorderRadius.circular(12),
                      border: Border.all(
                        color: _error != null ? Brand.errorStrong : (i == _code.length ? Brand.green : Brand.divider),
                        width: _error != null || i == _code.length ? 2 : 1,
                      ),
                    ),
                    child: Text(i < _code.length ? _code[i] : '', style: AppText.title.copyWith(fontSize: 26)),
                  ),
                ),
            ],
          ),
        ),
        if (_error != null)
          Padding(
            padding: const EdgeInsets.fromLTRB(4, 10, 4, 0),
            child: Text(_error!, style: AppText.label.copyWith(color: Brand.errorStrong)),
          ),
        const SizedBox(height: 16),
        NumberPad(onDigit: _busy ? null : _type, onBackspace: _back),
        const SizedBox(height: 16),
        BrandButton(text: 'Connect', loading: _busy, onPressed: _code.length == 6 ? _connect : null),
      ],
    );
  }
}

/// Forest band with the logo, like the rest of the setup screens.
class _Header extends StatelessWidget {
  const _Header({required this.title, required this.subtitle});

  final String title;
  final String subtitle;

  @override
  Widget build(BuildContext context) {
    return Container(
      color: Brand.forest,
      child: SafeArea(
        bottom: false,
        child: Padding(
          padding: const EdgeInsets.fromLTRB(20, 20, 20, 28),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  const BhoomyLogo(size: 30, tileColor: Brand.forestRaised),
                  const SizedBox(width: 10),
                  Text('Bhoomy POS', style: AppText.subhead.copyWith(color: Colors.white)),
                ],
              ),
              const SizedBox(height: 18),
              Text(title, style: AppText.title.copyWith(color: Colors.white, fontSize: 26)),
              const SizedBox(height: 4),
              Text(subtitle, style: AppText.body.copyWith(color: Brand.greenTint)),
            ],
          ),
        ),
      ),
    );
  }
}

class _Preparing extends ConsumerWidget {
  const _Preparing({required this.error, required this.onRetry});

  final String? error;
  final VoidCallback? onRetry;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final location = ref.watch(sessionEventProvider)?.name ?? '';
    return SectionCard(
      child: Column(
        children: [
          if (error == null) ...[
            const SizedBox(height: 8),
            const CircularProgressIndicator(),
            const SizedBox(height: 16),
            Text('Getting this device ready for', style: AppText.body.copyWith(color: Brand.slate)),
            Text(location, textAlign: TextAlign.center, style: AppText.heading),
            const SizedBox(height: 4),
            Text('Products, promotions and staff', style: AppText.caption),
          ] else ...[
            Text("Couldn't finish setup", style: AppText.bodyLarge.copyWith(color: Brand.errorInk)),
            const SizedBox(height: 6),
            Text(
              error!,
              textAlign: TextAlign.center,
              style: AppText.body.copyWith(color: Brand.errorInk),
            ),
            const SizedBox(height: 14),
            BrandButton(text: 'Try again', onPressed: onRetry),
          ],
        ],
      ),
    );
  }
}

class _Ready extends ConsumerWidget {
  const _Ready(this.counts);

  final ({int products, int promotions, int staff}) counts;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final device = ref.watch(deviceProvider);
    if (device == null) return const SizedBox.shrink();
    String n(int count, String noun) => '$count $noun${count == 1 ? '' : 's'}';
    final rows = [
      ('Device', device.name),
      ('Store', device.storeName),
      (
        'Data',
        '${n(counts.products, 'product')}, ${n(counts.promotions, 'promotion')}, ${counts.staff} staff downloaded',
      ),
    ];
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 28),
          decoration: BoxDecoration(
            color: Brand.surface,
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: Brand.cardLine),
          ),
          child: Column(
            children: [
              const CircleAvatar(
                radius: 32,
                backgroundColor: Brand.successSoft,
                child: Icon(Icons.check_rounded, size: 34, color: Brand.greenPressed),
              ),
              const SizedBox(height: 12),
              Text(
                'Connected to ${device.location.name}',
                textAlign: TextAlign.center,
                style: AppText.heading.copyWith(fontWeight: FontWeight.w900),
              ),
              const SizedBox(height: 8),
              Pill(device.location.placeLabel, color: Brand.goldInk, background: Brand.goldTint),
              const SizedBox(height: 16),
              for (final (label, value) in rows)
                Padding(
                  padding: const EdgeInsets.symmetric(vertical: 3),
                  child: Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      SizedBox(
                        width: 72,
                        child: Text(label, style: AppText.body.copyWith(color: Brand.muted)),
                      ),
                      Expanded(
                        child: Text(value, style: AppText.body.copyWith(fontWeight: FontWeight.w700)),
                      ),
                    ],
                  ),
                ),
            ],
          ),
        ),
        const SizedBox(height: 16),
        Text(
          "This device stays connected until it's removed in Bhoomy. No need to enter a code again.",
          textAlign: TextAlign.center,
          style: AppText.caption.copyWith(fontSize: 13),
        ),
        const SizedBox(height: 20),
        BrandButton(text: 'Continue', onPressed: () => context.go('/lock')),
      ],
    );
  }
}
