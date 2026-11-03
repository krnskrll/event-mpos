import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../app/logo.dart';
import '../../app/theme.dart';
import '../../app/widgets.dart';
import '../../core/network/api_failure.dart';
import '../../data/event/event_repository.dart';
import '../../data/session/session.dart';

/// Setup steps: what the cashier sees, and what is being loaded.
const _steps = ['Products and prices', 'Members', 'Promotions', 'Stock levels', 'Order numbers for this device'];

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
  ({int products, int members, int promotions})? _counts;

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
      final events = ref.read(eventRepositoryProvider);
      await Future.wait([events.syncSnapshot(event.id), _walkSteps()]);
      final counts = await events.localCounts(event.id);
      if (!mounted) return;
      setState(() {
        _done = true;
        _counts = counts;
      });
      await Future.delayed(const Duration(milliseconds: 900));
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
    final step = _done ? _steps.length : _step;
    final c = _counts;
    final details = [
      if (c != null) '${c.products} products',
      if (c != null) '${c.members} members',
      if (c != null) '${c.promotions} running',
    ];
    return Scaffold(
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.fromLTRB(24, 32, 24, 24),
            child: ConstrainedBox(
              constraints: const BoxConstraints(maxWidth: 440),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  const Align(alignment: Alignment.centerLeft, child: BhoomyLogo(size: 56)),
                  const SizedBox(height: 22),
                  Text('Getting this device ready for', style: AppText.body.copyWith(color: Brand.slate)),
                  Text(ref.watch(sessionEventProvider)?.name ?? '', style: AppText.title.copyWith(fontSize: 26)),
                  const SizedBox(height: 22),
                  Container(
                    padding: const EdgeInsets.all(20),
                    decoration: BoxDecoration(
                      color: Brand.surface,
                      borderRadius: BorderRadius.circular(24),
                      border: Border.all(color: Brand.cardLine),
                      boxShadow: AppShadows.card,
                    ),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: [
                        Row(
                          children: [
                            Expanded(
                              child: Text(
                                _error != null
                                    ? "Couldn't finish setup"
                                    : (_done ? 'All set' : 'Loading ${_steps[step].toLowerCase()}'),
                                style: AppText.bodyLarge.copyWith(
                                  color: _error != null ? Brand.errorInk : Brand.forest,
                                ),
                              ),
                            ),
                            Text(
                              '${(step + 1).clamp(1, _steps.length)} of ${_steps.length}',
                              style: AppText.label.copyWith(color: Brand.green),
                            ),
                          ],
                        ),
                        const SizedBox(height: 14),
                        LinearProgressIndicator(
                          value: _done ? 1 : (step + 0.5) / _steps.length,
                          minHeight: 8,
                          borderRadius: BorderRadius.circular(999),
                        ),
                        const SizedBox(height: 8),
                        for (final (i, label) in _steps.indexed)
                          _StepRow(
                            label,
                            detail: i < details.length ? details[i] : null,
                            state: i < step
                                ? _Step.done
                                : (i == step && _error == null ? _Step.current : _Step.waiting),
                          ),
                        if (_error != null) ...[
                          const SizedBox(height: 8),
                          Text(_error!, style: AppText.body.copyWith(color: Brand.errorInk)),
                          const SizedBox(height: 14),
                          BrandButton(text: 'Try again', onPressed: _retry),
                        ],
                      ],
                    ),
                  ),
                  const SizedBox(height: 18),
                  Container(
                    padding: const EdgeInsets.all(14),
                    decoration: BoxDecoration(color: Brand.greenSoft, borderRadius: BorderRadius.circular(14)),
                    child: Row(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Icon(Icons.info_outline, size: 18, color: Brand.forest),
                        const SizedBox(width: 10),
                        Expanded(
                          child: Text(
                            'Keep the app open. After this, the device can keep selling even without internet.',
                            style: AppText.caption.copyWith(color: Brand.forest, fontSize: 13),
                          ),
                        ),
                      ],
                    ),
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

enum _Step { done, current, waiting }

class _StepRow extends StatelessWidget {
  const _StepRow(this.label, {required this.state, this.detail});

  final String label;
  final _Step state;
  final String? detail;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 8),
      child: Row(
        children: [
          Container(
            width: 24,
            height: 24,
            decoration: BoxDecoration(
              shape: BoxShape.circle,
              color: switch (state) {
                _Step.done => Brand.green,
                _Step.current => Brand.surface,
                _Step.waiting => Brand.sunken,
              },
              border: state == _Step.current ? Border.all(color: Brand.green, width: 3) : null,
            ),
            child: state == _Step.done ? const Icon(Icons.check, size: 14, color: Colors.white) : null,
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Text(
              label,
              style: AppText.body.copyWith(
                color: state == _Step.waiting ? Brand.faint : Brand.ink,
                fontWeight: state == _Step.current ? FontWeight.w700 : FontWeight.w400,
              ),
            ),
          ),
          if (detail != null) Text(detail!, style: AppText.caption),
        ],
      ),
    );
  }
}
