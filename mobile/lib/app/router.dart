import 'package:flutter/widgets.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../core/secure/token_storage.dart';
import '../data/session/session.dart';
import '../features/cart/cart_screen.dart';
import '../features/catalogue/catalogue_screen.dart';
import '../features/catalogue/product_detail_screen.dart';
import '../features/customer/customer_screen.dart';
import '../features/events/event_selection_screen.dart';
import '../features/home/home_screen.dart';
import '../features/login/login_screen.dart';
import '../features/orders/orders_screen.dart';
import '../features/printer/printer_screen.dart';
import '../features/receipt/receipt_screen.dart';
import '../features/scan/barcode_scan_screen.dart';
import '../features/sync/sync_screen.dart';
import 'shell.dart';

GoRoute _route(String path, Widget Function(GoRouterState state) build) =>
    GoRoute(path: path, builder: (_, state) => build(state));

/// Signed out → /login. Signed in without a selected event → /events. Sign-in changes re-run this.
/// The four tabs live in the shell; scanners, product detail, receipt and printer open full screen.
final routerProvider = Provider<GoRouter>((ref) {
  final tokens = ref.watch(tokenStorageProvider);
  const open = {'/login', '/events'};
  return GoRouter(
    initialLocation: '/events',
    refreshListenable: tokens,
    redirect: (_, state) {
      final at = state.matchedLocation;
      if (!tokens.isLoggedIn) return at == '/login' ? null : '/login';
      if (at == '/login' || (!open.contains(at) && ref.read(sessionEventProvider) == null)) return '/events';
      return null;
    },
    routes: [
      _route('/login', (_) => const LoginScreen()),
      _route('/events', (_) => const EventSelectionScreen()),
      _route('/sync', (_) => const SyncScreen()),
      StatefulShellRoute.indexedStack(
        builder: (_, _, shell) => AppShell(shell),
        branches: [
          for (final (path, screen) in [
            ('/home', const HomeScreen()),
            ('/catalogue', const CatalogueScreen()),
            ('/cart', const CartScreen()),
            ('/orders', const OrdersScreen()),
          ])
            StatefulShellBranch(routes: [_route(path, (_) => screen)]),
        ],
      ),
      _route('/product/:id', (s) => ProductDetailScreen(s.pathParameters['id']!)),
      _route('/scan', (_) => const BarcodeScanScreen()),
      _route('/customer', (_) => const CustomerScreen()),
      _route('/receipt/:key', (s) => ReceiptScreen(s.pathParameters['key']!)),
      _route('/printer', (_) => const PrinterScreen()),
    ],
  );
});
