import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../app/theme.dart';
import '../../app/widgets.dart';
import '../../core/secure/token_storage.dart';
import '../../data/auth/auth_repository.dart';
import '../../data/event/event_repository.dart';
import '../../data/session/session.dart';
import '../../domain/event.dart';

final _eventsProvider = FutureProvider.autoDispose((ref) => ref.watch(eventRepositoryProvider).assignedEvents());

class EventSelectionScreen extends ConsumerWidget {
  const EventSelectionScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    void start(AssignedEvent event) {
      ref.read(sessionEventProvider.notifier).start(event);
      context.go('/sync');
    }

    final name = ref.watch(tokenStorageProvider).userName ?? 'Cashier';
    return Scaffold(
      body: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          ScreenHeader(
            title: 'Hi, ${name.split(' ').first}',
            subtitle: 'Choose the event you are selling at',
            actions: [
              TextButton(onPressed: () => ref.read(authRepositoryProvider).logout(), child: const Text('Log out')),
            ],
          ),
          Expanded(
            child: ref
                .watch(_eventsProvider)
                .when(
                  loading: () => const Center(child: CircularProgressIndicator()),
                  error: (e, _) => Padding(
                    padding: const EdgeInsets.all(20),
                    child: Column(
                      children: [
                        Text('$e', style: AppText.body.copyWith(color: Brand.errorInk)),
                        const SizedBox(height: 12),
                        BrandButton(text: 'Retry', onPressed: () => ref.invalidate(_eventsProvider)),
                      ],
                    ),
                  ),
                  data: (events) => events.isEmpty
                      ? const EmptyState('No events assigned to you yet.', icon: Icons.event_busy_outlined)
                      : ListView.separated(
                          padding: const EdgeInsets.fromLTRB(16, 8, 16, 16),
                          itemCount: events.length,
                          separatorBuilder: (_, _) => const SizedBox(height: 12),
                          itemBuilder: (_, i) => _EventCard(events[i], onStart: () => start(events[i])),
                        ),
                ),
          ),
        ],
      ),
    );
  }
}

class _EventCard extends StatelessWidget {
  const _EventCard(this.event, {required this.onStart});

  final AssignedEvent event;
  final VoidCallback onStart;

  @override
  Widget build(BuildContext context) {
    final active = event.status == EventStatus.active;
    return SectionCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          active
              ? const Pill('Active', color: Brand.successInk, background: Brand.successSoft)
              : Pill(
                  event.status.name[0].toUpperCase() + event.status.name.substring(1),
                  color: Brand.warningInk,
                  background: Brand.warningSoft,
                ),
          const SizedBox(height: 10),
          Text(event.name, style: AppText.subhead),
          const SizedBox(height: 6),
          _Meta(Icons.calendar_month_outlined, '${event.startDate ?? '—'} → ${event.endDate ?? '—'}'),
          _Meta(Icons.place_outlined, event.code),
          const SizedBox(height: 16),
          BrandButton(text: active ? 'Start selling' : 'Not active yet', onPressed: active ? onStart : null),
        ],
      ),
    );
  }
}

class _Meta extends StatelessWidget {
  const _Meta(this.icon, this.text);

  final IconData icon;
  final String text;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(top: 2),
      child: Row(
        children: [
          Icon(icon, color: Brand.muted, size: 15),
          const SizedBox(width: 6),
          Text(text, style: AppText.caption),
        ],
      ),
    );
  }
}
