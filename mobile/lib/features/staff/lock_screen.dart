import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../app/theme.dart';
import '../../app/widgets.dart';
import '../../core/connectivity.dart';
import '../../data/device/device_repository.dart';
import '../../data/session/session.dart';
import '../../domain/event.dart';
import '../../domain/staff.dart';
import 'pin_pad.dart';

/// "Who's selling?": pick your name, type your PIN. The PIN is checked on the device against the
/// cached hash, so signing in works offline.
class LockScreen extends ConsumerStatefulWidget {
  const LockScreen({super.key});

  @override
  ConsumerState<LockScreen> createState() => _LockScreenState();
}

class _LockScreenState extends ConsumerState<LockScreen> {
  StaffMember? _picked;
  String _pin = '';
  String? _error;
  bool _checking = false;
  Duration? _wait;
  Timer? _ticker;

  @override
  void dispose() {
    _ticker?.cancel();
    super.dispose();
  }

  void _pick(StaffMember? staff) {
    setState(() {
      _picked = staff;
      _pin = '';
      _error = null;
    });
    _tick();
  }

  /// Lockout countdown for the picked staff member, once a second until it runs out.
  void _tick() {
    _ticker?.cancel();
    final staff = _picked;
    setState(() => _wait = staff == null ? null : ref.read(pinAttemptsProvider).waitFor(staff.id));
    if (_wait != null) _ticker = Timer(const Duration(seconds: 1), _tick);
  }

  void _type(String digit) => setState(() {
    if (_pin.length < 6) _pin += digit;
    _error = null;
  });

  void _back() => setState(() {
    if (_pin.isNotEmpty) _pin = _pin.substring(0, _pin.length - 1);
  });

  Future<void> _submit() async {
    final staff = _picked!;
    setState(() => _checking = true);
    final ok = await ref.read(pinCheckProvider)(_pin, staff.pinHash);
    if (!mounted) return;
    final attempts = ref.read(pinAttemptsProvider);
    if (ok) {
      attempts.reset(staff.id);
      ref.read(signedInProvider.notifier).signIn(staff);
      return;
    }
    final left = attempts.fail(staff.id);
    setState(() {
      _checking = false;
      _pin = '';
      _error = left > 0 ? 'Wrong PIN. $left ${left == 1 ? 'try' : 'tries'} left.' : null;
    });
    _tick();
  }

  @override
  Widget build(BuildContext context) {
    final picked = _picked;
    return AnnotatedRegion(
      value: SystemUiOverlayStyle.light.copyWith(statusBarColor: Colors.transparent),
      child: PopScope(
        canPop: false,
        onPopInvokedWithResult: (_, _) => _pick(null),
        child: Scaffold(
          body: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              _Header(onBack: picked == null ? null : () => _pick(null)),
              Expanded(
                child: Center(
                  child: SingleChildScrollView(
                    padding: const EdgeInsets.all(20),
                    child: ConstrainedBox(
                      constraints: BoxConstraints(maxWidth: picked == null ? 760 : 420),
                      child: picked == null ? _StaffGrid(onPick: _pick) : _pinEntry(picked),
                    ),
                  ),
                ),
              ),
              const _Footer(),
            ],
          ),
        ),
      ),
    );
  }

  Widget _pinEntry(StaffMember staff) {
    final wait = _wait;
    final message = wait != null ? 'Too many wrong tries. Try again in ${countdown(wait)}.' : _error;
    return Column(
      children: [
        StaffAvatar(staff, radius: 32),
        const SizedBox(height: 12),
        Text(
          staff.name,
          textAlign: TextAlign.center,
          style: AppText.heading.copyWith(fontWeight: FontWeight.w900),
        ),
        Text('Enter your PIN', style: AppText.body.copyWith(color: Brand.muted)),
        const SizedBox(height: 16),
        PinDots(_pin.length, error: message != null),
        const SizedBox(height: 10),
        SizedBox(
          height: 22,
          child: message == null
              ? null
              : Text(
                  message,
                  style: AppText.label.copyWith(color: Brand.errorStrong),
                  textAlign: TextAlign.center,
                ),
        ),
        const SizedBox(height: 8),
        NumberPad(onDigit: wait != null || _checking ? null : _type, onBackspace: _back),
        const SizedBox(height: 12),
        BrandButton(text: 'Sign in', loading: _checking, onPressed: _pin.length >= 4 && wait == null ? _submit : null),
      ],
    );
  }
}

/// Location name, what kind of place it is, and the time.
class _Header extends ConsumerStatefulWidget {
  const _Header({this.onBack});

  final VoidCallback? onBack;

  @override
  ConsumerState<_Header> createState() => _HeaderState();
}

class _HeaderState extends ConsumerState<_Header> {
  late final Timer _clock;

  @override
  void initState() {
    super.initState();
    _clock = Timer.periodic(const Duration(seconds: 20), (_) => setState(() {}));
  }

  @override
  void dispose() {
    _clock.cancel();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final location = ref.watch(sessionEventProvider);
    return Container(
      color: Brand.forest,
      child: SafeArea(
        bottom: false,
        child: Padding(
          padding: const EdgeInsets.fromLTRB(12, 14, 20, 18),
          child: Row(
            children: [
              if (widget.onBack != null)
                IconButton(
                  onPressed: widget.onBack,
                  tooltip: 'Back to staff list',
                  color: Colors.white,
                  icon: const Icon(Icons.arrow_back),
                )
              else
                const SizedBox(width: 8),
              const SizedBox(width: 4),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      location?.name ?? '',
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: AppText.heading.copyWith(color: Colors.white, fontWeight: FontWeight.w900),
                    ),
                    Text(
                      location?.placeLabel ?? '',
                      style: AppText.caption.copyWith(color: Brand.greenTint, fontSize: 13),
                    ),
                  ],
                ),
              ),
              Text(
                TimeOfDay.now().format(context),
                style: AppText.heading.copyWith(
                  color: Colors.white,
                  fontFeatures: const [FontFeature.tabularFigures()],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _StaffGrid extends ConsumerWidget {
  const _StaffGrid({required this.onPick});

  final ValueChanged<StaffMember> onPick;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final staff = ref.watch(staffProvider).value;
    if (staff == null) return const Center(child: CircularProgressIndicator());
    if (staff.isEmpty) {
      return const EmptyState(
        'No staff for this location yet.\nAsk the store owner to add you in Bhoomy.',
        icon: Icons.group_off_outlined,
      );
    }
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Text("Who's selling?", style: AppText.subhead.copyWith(fontWeight: FontWeight.w900)),
        const SizedBox(height: 14),
        GridView(
          shrinkWrap: true,
          physics: const NeverScrollableScrollPhysics(),
          gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
            crossAxisCount: isWide(context) ? 4 : 2,
            mainAxisSpacing: 12,
            crossAxisSpacing: 12,
            mainAxisExtent: 140,
          ),
          children: [
            for (final s in staff)
              Material(
                color: Brand.surface,
                borderRadius: BorderRadius.circular(14),
                child: InkWell(
                  borderRadius: BorderRadius.circular(14),
                  onTap: () => onPick(s),
                  child: Padding(
                    padding: const EdgeInsets.all(12),
                    child: Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        StaffAvatar(s),
                        const SizedBox(height: 8),
                        Text(
                          s.name,
                          maxLines: 2,
                          overflow: TextOverflow.ellipsis,
                          textAlign: TextAlign.center,
                          style: AppText.body.copyWith(fontWeight: FontWeight.w700, height: 1.2),
                        ),
                        Text(s.roleLabel, style: AppText.caption),
                      ],
                    ),
                  ),
                ),
              ),
          ],
        ),
        const SizedBox(height: 16),
        Text('Name missing? Ask the store owner to add you in Bhoomy.', style: AppText.caption.copyWith(fontSize: 13)),
      ],
    );
  }
}

/// Which device this is, and whether it is online.
class _Footer extends ConsumerWidget {
  const _Footer();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final online = ref.isOnline;
    return Container(
      decoration: const BoxDecoration(
        color: Brand.surface,
        border: Border(top: BorderSide(color: Brand.divider)),
      ),
      child: SafeArea(
        top: false,
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 14),
          child: Row(
            children: [
              Expanded(
                child: Text.rich(
                  TextSpan(
                    children: [
                      const TextSpan(text: 'Device: '),
                      TextSpan(
                        text: ref.watch(deviceProvider)?.name ?? '',
                        style: const TextStyle(fontWeight: FontWeight.w700, color: Brand.ink),
                      ),
                    ],
                  ),
                  style: AppText.caption.copyWith(fontSize: 13),
                ),
              ),
              Icon(Icons.circle, size: 8, color: online ? Brand.success : Brand.warning),
              const SizedBox(width: 6),
              Text(online ? 'Online' : 'Offline', style: AppText.caption.copyWith(fontSize: 13)),
            ],
          ),
        ),
      ),
    );
  }
}
