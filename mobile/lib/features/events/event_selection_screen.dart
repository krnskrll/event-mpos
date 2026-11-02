import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../app/logo.dart';
import '../../app/status.dart';
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
    final wide = isWide(context);
    return Scaffold(
      body: SafeArea(
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 1000),
            child: Padding(
              padding: EdgeInsets.fromLTRB(16, wide ? 28 : 16, 16, 0),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  Row(
                    children: [
                      wide ? const BhoomyLogo(size: 44) : const UserAvatar(),
                      const SizedBox(width: 12),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text('Hi, ${name.split(' ').first}', style: AppText.title),
                            Text("Choose where you're selling today", style: AppText.body.copyWith(color: Brand.slate)),
                          ],
                        ),
                      ),
                      wide
                          ? OutlinedButton(
                              onPressed: () => ref.read(authRepositoryProvider).logout(),
                              style: OutlinedButton.styleFrom(minimumSize: const Size(0, 44)),
                              child: const Text('Log out'),
                            )
                          : TextButton(
                              onPressed: () => ref.read(authRepositoryProvider).logout(),
                              child: const Text('Log out'),
                            ),
                      if (wide) ...[const SizedBox(width: 12), const UserAvatar()],
                    ],
                  ),
                  const SizedBox(height: 20),
                  Expanded(
                    child: ref
                        .watch(_eventsProvider)
                        .when(
                          loading: () => const Center(child: CircularProgressIndicator()),
                          error: (e, _) => Column(
                            children: [
                              Text('$e', style: AppText.body.copyWith(color: Brand.errorInk)),
                              const SizedBox(height: 12),
                              BrandButton(text: 'Try again', onPressed: () => ref.invalidate(_eventsProvider)),
                            ],
                          ),
                          data: (events) => events.isEmpty
                              ? const EmptyState(
                                  'No events assigned to you yet.\nAsk your supervisor to add you to one.',
                                  icon: Icons.event_busy_outlined,
                                )
                              : GridView.builder(
                                  padding: const EdgeInsets.only(bottom: 16),
                                  gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
                                    crossAxisCount: wide ? 2 : 1,
                                    mainAxisSpacing: 16,
                                    crossAxisSpacing: 18,
                                    mainAxisExtent: wide ? 268 : 232,
                                  ),
                                  itemCount: events.length,
                                  itemBuilder: (_, i) => _EventCard(events[i], onStart: () => start(events[i])),
                                ),
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

/// Coloured top with status and code; name, dates, and the way in (only open events can start).
class _EventCard extends StatelessWidget {
  const _EventCard(this.event, {required this.onStart});

  final AssignedEvent event;
  final VoidCallback onStart;

  @override
  Widget build(BuildContext context) {
    final now = DateTime.now();
    final (status, band, shape, pillBg, pillFg, codeColor) = switch (event.status) {
      EventStatus.active => ('Open now', Brand.forest, Brand.forestRaised, Brand.gold, Brand.onGold, Brand.greenTint),
      EventStatus.draft => ('Upcoming', Brand.greenSoft, Brand.greenPale, Brand.goldSoft, Brand.goldInk, Brand.forest),
      EventStatus.closed => ('Closed', Brand.sunken, Brand.divider, Brand.surface, Brand.slate, Brand.slate),
    };
    final (note, noteColor) = switch (event.status) {
      EventStatus.active => (event.dayOf(now) ?? 'Open', Brand.successInk),
      EventStatus.draft => (event.opensIn(now) ?? 'Not open yet', Brand.goldInk),
      EventStatus.closed => ('Ended', Brand.slate),
    };
    final open = event.status == EventStatus.active;
    return Container(
      clipBehavior: Clip.antiAlias,
      decoration: BoxDecoration(
        color: Brand.surface,
        borderRadius: BorderRadius.circular(24),
        border: Border.all(color: Brand.cardLine),
        boxShadow: AppShadows.card,
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Container(
            height: 84,
            decoration: BoxDecoration(color: band),
            clipBehavior: Clip.hardEdge, // keep the decorative circle inside the band
            padding: const EdgeInsets.all(16),
            child: Stack(
              clipBehavior: Clip.none,
              children: [
                Positioned(right: -56, top: -76, child: CircleAvatar(radius: 85, backgroundColor: shape)),
                Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Pill(status, color: pillFg, background: pillBg),
                    const Spacer(),
                    Text(event.code, style: AppText.label.copyWith(color: codeColor)),
                  ],
                ),
              ],
            ),
          ),
          Expanded(
            child: Padding(
              padding: const EdgeInsets.fromLTRB(16, 14, 16, 16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  Text(
                    event.name,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: AppText.heading.copyWith(fontWeight: FontWeight.w900),
                  ),
                  const SizedBox(height: 4),
                  Text.rich(
                    TextSpan(
                      children: [
                        TextSpan(text: '${event.dateRange} · '),
                        TextSpan(
                          text: note,
                          style: TextStyle(color: noteColor, fontWeight: FontWeight.w700),
                        ),
                      ],
                    ),
                    style: AppText.caption.copyWith(color: Brand.slate, fontSize: 13),
                  ),
                  const Spacer(),
                  BrandButton(text: open ? 'Start selling' : 'Not open yet', onPressed: open ? onStart : null),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}
