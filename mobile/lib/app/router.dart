import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

final routerProvider = Provider<GoRouter>((ref) {
  return GoRouter(
    routes: [
      GoRoute(path: '/', builder: (_, _) => const _StartScreen()),
    ],
  );
});

// Placeholder until the login flow lands in Phase 2.
class _StartScreen extends StatelessWidget {
  const _StartScreen();

  @override
  Widget build(BuildContext context) {
    return const Scaffold(
      body: Center(child: Text('event-mpos', style: TextStyle(fontSize: 24, fontWeight: FontWeight.bold))),
    );
  }
}
