import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
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

    return Scaffold(
      backgroundColor: Brand.bg,
      body: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _Header(
            name: ref.watch(tokenStorageProvider).userName ?? 'Beauty Advisor',
            onLogout: () => ref.read(authRepositoryProvider).logout(),
          ),
          const Padding(
            padding: EdgeInsets.fromLTRB(20, 20, 20, 8),
            child: Text(
              'YOUR ASSIGNED EVENTS',
              style: TextStyle(color: Brand.textFaint, fontSize: 13, fontWeight: FontWeight.w500, letterSpacing: 1),
            ),
          ),
          Expanded(
            child: ref
                .watch(_eventsProvider)
                .when(
                  loading: () => const Center(child: CircularProgressIndicator(color: Brand.navy)),
                  error: (e, _) => Padding(
                    padding: const EdgeInsets.all(20),
                    child: Column(
                      children: [
                        Text('$e', style: const TextStyle(color: Brand.danger)),
                        const SizedBox(height: 12),
                        BrandButton(text: 'Retry', onPressed: () => ref.invalidate(_eventsProvider)),
                      ],
                    ),
                  ),
                  data: (events) => events.isEmpty
                      ? const Padding(
                          padding: EdgeInsets.all(20),
                          child: Text('No events assigned to you yet.', style: TextStyle(color: Brand.textMuted)),
                        )
                      : ListView.separated(
                          padding: const EdgeInsets.all(16),
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

class _Header extends StatelessWidget {
  const _Header({required this.name, required this.onLogout});

  final String name;
  final VoidCallback onLogout;

  @override
  Widget build(BuildContext context) {
    final initials = name.split(' ').where((w) => w.isNotEmpty).take(2).map((w) => w[0].toUpperCase()).join();
    return AnnotatedRegion(
      value: SystemUiOverlayStyle.light, // light status-bar icons on navy
      child: Container(
        color: Brand.navy,
        padding: EdgeInsets.fromLTRB(20, MediaQuery.paddingOf(context).top + 16, 20, 20),
        child: Row(
          children: [
            CircleAvatar(
              radius: 20,
              backgroundColor: Brand.teal,
              child: Text(
                initials,
                style: const TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 14),
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text('Welcome,', style: TextStyle(color: Colors.white60, fontSize: 13)),
                  Text(
                    name,
                    style: const TextStyle(color: Colors.white, fontSize: 16, fontWeight: FontWeight.bold),
                  ),
                ],
              ),
            ),
            TextButton(
              onPressed: onLogout,
              style: TextButton.styleFrom(foregroundColor: Colors.white70, backgroundColor: Colors.white12),
              child: const Text('Logout'),
            ),
          ],
        ),
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
              ? const Pill('● ACTIVE', color: Brand.ready, background: Brand.readyBg)
              : Pill(event.status.name.toUpperCase(), color: Brand.packing, background: Brand.packingBg),
          const SizedBox(height: 10),
          Text(
            event.name,
            style: const TextStyle(color: Brand.heading, fontSize: 18, fontWeight: FontWeight.bold),
          ),
          const SizedBox(height: 6),
          _Meta(Icons.calendar_month_outlined, '${event.startDate ?? '—'} → ${event.endDate ?? '—'}'),
          _Meta(Icons.place_outlined, event.code),
          const SizedBox(height: 14),
          BrandButton(
            text: active ? 'Start Session' : 'Not active yet',
            onPressed: active ? onStart : null,
            color: Brand.teal,
          ),
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
          Icon(icon, color: Brand.textFaint, size: 15),
          const SizedBox(width: 6),
          Text(text, style: const TextStyle(color: Brand.textMuted, fontSize: 13)),
        ],
      ),
    );
  }
}
