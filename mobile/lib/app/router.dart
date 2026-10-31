import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../core/secure/token_storage.dart';
import '../data/auth/auth_repository.dart';
import '../data/session/session.dart';
import '../features/catalogue/catalogue_screen.dart';
import '../features/catalogue/product_detail_screen.dart';
import '../features/events/event_selection_screen.dart';
import '../features/login/login_screen.dart';
import '../features/sync/sync_screen.dart';
import 'widgets.dart';

/// Signed out → /login. Signed in without a selected event → /events. Sign-in changes re-run this.
final routerProvider = Provider<GoRouter>((ref) {
  final tokens = ref.watch(tokenStorageProvider);
  return GoRouter(
    initialLocation: '/events',
    refreshListenable: tokens,
    redirect: (_, state) {
      final at = state.matchedLocation;
      if (!tokens.isLoggedIn) return at == '/login' ? null : '/login';
      if (at == '/login' || (at != '/events' && ref.read(sessionEventProvider) == null)) return '/events';
      return null;
    },
    routes: [
      GoRoute(path: '/login', builder: (_, _) => const LoginScreen()),
      GoRoute(path: '/events', builder: (_, _) => const EventSelectionScreen()),
      GoRoute(path: '/sync', builder: (_, _) => const SyncScreen()),
      GoRoute(path: '/home', builder: (_, _) => const _HomePlaceholder()),
      GoRoute(
        path: '/catalogue',
        builder: (_, _) => const CatalogueScreen(),
        routes: [GoRoute(path: ':id', builder: (_, state) => ProductDetailScreen(state.pathParameters['id']!))],
      ),
    ],
  );
});

// Placeholder until the home screen and bottom bar land in Phase 9.
class _HomePlaceholder extends ConsumerWidget {
  const _HomePlaceholder();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return Scaffold(
      body: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Text(
              ref.watch(sessionEventProvider)?.name ?? '',
              style: const TextStyle(fontSize: 22, fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 24),
            BrandButton(text: 'Browse products', onPressed: () => context.push('/catalogue')),
            const SizedBox(height: 12),
            BrandButton(text: 'Sign Out', onPressed: () => ref.read(authRepositoryProvider).logout()),
          ],
        ),
      ),
    );
  }
}
