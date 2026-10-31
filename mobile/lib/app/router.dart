import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../core/secure/token_storage.dart';
import '../data/auth/auth_repository.dart';
import '../features/login/login_screen.dart';
import 'widgets.dart';

/// Signed-out users always land on /login; signing in or out re-runs the redirect.
final routerProvider = Provider<GoRouter>((ref) {
  final tokens = ref.watch(tokenStorageProvider);
  return GoRouter(
    refreshListenable: tokens,
    redirect: (_, state) {
      final atLogin = state.matchedLocation == '/login';
      if (!tokens.isLoggedIn) return atLogin ? null : '/login';
      return atLogin ? '/' : null;
    },
    routes: [
      GoRoute(path: '/login', builder: (_, _) => const LoginScreen()),
      GoRoute(path: '/', builder: (_, _) => const _SignedInScreen()),
    ],
  );
});

// Placeholder until event selection lands in Phase 3.
class _SignedInScreen extends ConsumerWidget {
  const _SignedInScreen();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return Scaffold(
      body: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Text('Signed in', style: TextStyle(fontSize: 24, fontWeight: FontWeight.bold)),
            const SizedBox(height: 24),
            BrandButton(text: 'Sign Out', onPressed: () => ref.read(authRepositoryProvider).logout()),
          ],
        ),
      ),
    );
  }
}
