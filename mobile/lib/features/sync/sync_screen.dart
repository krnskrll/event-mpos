import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../app/theme.dart';
import '../../app/widgets.dart';
import '../../core/network/api_failure.dart';
import '../../data/event/event_repository.dart';
import '../../data/session/session.dart';

const _steps = [
  'Downloading products…',
  'Syncing member data…',
  'Loading promotions…',
  'Fetching inventory…',
  'Finalising setup…',
];

/// Downloads the event snapshot while walking through the setup steps, then opens the session.
class SyncScreen extends ConsumerStatefulWidget {
  const SyncScreen({super.key});

  @override
  ConsumerState<SyncScreen> createState() => _SyncScreenState();
}

class _SyncScreenState extends ConsumerState<SyncScreen> {
  int _step = 0;
  bool _done = false;
  String? _error;

  @override
  void initState() {
    super.initState();
    _run();
  }

  void _retry() {
    setState(() {
      _step = 0;
      _error = null;
    });
    _run();
  }

  Future<void> _run() async {
    final event = ref.read(sessionEventProvider)!;
    try {
      await Future.wait([ref.read(eventRepositoryProvider).syncSnapshot(event.id), _walkSteps()]);
      if (!mounted) return;
      setState(() => _done = true);
      await Future.delayed(const Duration(milliseconds: 700));
      if (mounted) context.go('/home');
    } on ApiFailure catch (e) {
      if (mounted) setState(() => _error = e.message);
    }
  }

  Future<void> _walkSteps() async {
    for (var i = 1; i <= _steps.length; i++) {
      await Future.delayed(const Duration(milliseconds: 480));
      if (!mounted || i == _steps.length) return;
      setState(() => _step = i);
    }
  }

  @override
  Widget build(BuildContext context) {
    final progress = _done ? 1.0 : (_step + 1) / _steps.length;
    final message = _error ?? (_done ? 'All done! Starting session…' : _steps[_step]);
    return Scaffold(
      backgroundColor: Colors.white,
      body: Center(
        child: SingleChildScrollView(
          padding: const EdgeInsets.symmetric(horizontal: 32),
          child: Column(
            children: [
              const Text(
                'SETTING UP EVENT',
                style: TextStyle(color: Brand.textFaint, fontSize: 13, fontWeight: FontWeight.w500, letterSpacing: 1),
              ),
              const SizedBox(height: 6),
              Text(
                ref.watch(sessionEventProvider)?.name ?? '',
                textAlign: TextAlign.center,
                style: const TextStyle(color: Brand.heading, fontSize: 20, fontWeight: FontWeight.bold),
              ),
              const SizedBox(height: 32),
              CircleAvatar(
                radius: 44,
                backgroundColor: _done ? Brand.tealSoft : Brand.navActiveBg,
                child: Text(
                  _done ? '✓' : '${(progress * 100).round()}%',
                  style: TextStyle(
                    color: _done ? Brand.teal : Brand.navy,
                    fontSize: _done ? 40 : 18,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ),
              Padding(
                padding: const EdgeInsets.symmetric(vertical: 24),
                child: Text(
                  message,
                  textAlign: TextAlign.center,
                  style: TextStyle(color: _error != null ? Brand.danger : Brand.textMuted, fontSize: 15),
                ),
              ),
              LinearProgressIndicator(
                value: progress,
                minHeight: 6,
                color: Brand.teal,
                backgroundColor: Brand.cardBorder,
                borderRadius: BorderRadius.circular(999),
              ),
              const SizedBox(height: 28),
              for (final (i, step) in _steps.indexed)
                _StepRow(step, done: _done || i < _step, current: !_done && i == _step),
              if (_error != null) ...[const SizedBox(height: 24), BrandButton(text: 'Retry', onPressed: _retry)],
            ],
          ),
        ),
      ),
    );
  }
}

class _StepRow extends StatelessWidget {
  const _StepRow(this.label, {required this.done, required this.current});

  final String label;
  final bool done;
  final bool current;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        children: [
          CircleAvatar(
            radius: 10,
            backgroundColor: done ? Brand.teal : (current ? Brand.navActiveBg : Brand.cardBorder),
            child: done ? const Icon(Icons.check, color: Colors.white, size: 12) : null,
          ),
          const SizedBox(width: 12),
          Text(
            label,
            style: TextStyle(
              color: done ? Brand.teal : (current ? Brand.heading : Brand.textGhost),
              fontSize: 13,
              fontWeight: current ? FontWeight.w600 : FontWeight.normal,
            ),
          ),
        ],
      ),
    );
  }
}
