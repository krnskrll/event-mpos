import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../app/logo.dart';
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
      body: Center(
        child: SingleChildScrollView(
          padding: const EdgeInsets.symmetric(horizontal: 32),
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 420),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const BhoomyLogo(size: 48),
                const SizedBox(height: 32),
                Text('Setting up', style: AppText.caption.copyWith(color: Brand.slate)),
                Text(ref.watch(sessionEventProvider)?.name ?? '', style: AppText.title),
                const SizedBox(height: 24),
                LinearProgressIndicator(value: progress, minHeight: 6, borderRadius: BorderRadius.circular(999)),
                Padding(
                  padding: const EdgeInsets.symmetric(vertical: 16),
                  child: Text(
                    message,
                    style: AppText.body.copyWith(color: _error != null ? Brand.errorInk : Brand.slate),
                  ),
                ),
                for (final (i, step) in _steps.indexed)
                  _StepRow(step, done: _done || i < _step, current: !_done && i == _step),
                if (_error != null) ...[const SizedBox(height: 24), BrandButton(text: 'Retry', onPressed: _retry)],
              ],
            ),
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
      padding: const EdgeInsets.symmetric(vertical: 6),
      child: Row(
        children: [
          CircleAvatar(
            radius: 10,
            backgroundColor: done ? Brand.green : (current ? Brand.greenSoft : Brand.sunken),
            child: done ? const Icon(Icons.check, color: Colors.white, size: 12) : null,
          ),
          const SizedBox(width: 12),
          Text(
            label,
            style: AppText.body.copyWith(
              color: done || current ? Brand.ink : Brand.faint,
              fontWeight: current ? FontWeight.w700 : FontWeight.w400,
            ),
          ),
        ],
      ),
    );
  }
}
