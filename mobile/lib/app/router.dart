import 'package:flutter/widgets.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../core/secure/device_storage.dart';
import '../data/session/session.dart';
import '../domain/staff.dart';
import '../features/cart/cart_screen.dart';
import '../features/customer/customer_screen.dart';
import '../features/device/pair_screen.dart';
import '../features/home/home_screen.dart';
import '../features/orders/order_detail.dart';
import '../features/orders/orders_screen.dart';
import '../features/printer/printer_screen.dart';
import '../features/receipt/receipt_screen.dart';
import '../features/scan/barcode_scan_screen.dart';
import '../features/sell/sell_screen.dart';
import '../features/staff/lock_screen.dart';
import 'shell.dart';

GoRoute _route(String path, Widget Function(GoRouterState state) build) =>
    GoRoute(path: path, builder: (_, state) => build(state));

/// Not paired → /pair. Paired but nobody signed in → /lock. Pairing and sign-in changes re-run this.
/// The tabs live in the shell; the phone cart, scanners, receipt and printer open full screen.
final routerProvider = Provider<GoRouter>((ref) {
  final storage = ref.watch(deviceStorageProvider);
  final signedIn = ValueNotifier<StaffMember?>(null);
  ref.listen(signedInProvider, (_, staff) => signedIn.value = staff);
  ref.onDispose(signedIn.dispose);
  return GoRouter(
    initialLocation: '/home',
    refreshListenable: Listenable.merge([storage, signedIn]),
    redirect: (_, state) {
      final at = state.matchedLocation;
      if (!storage.isPaired) return at == '/pair' ? null : '/pair';
      if (at == '/pair') return null; // finishes its download, then continues to /lock
      if (ref.read(signedInProvider) == null) return at == '/lock' ? null : '/lock';
      return at == '/lock' ? '/home' : null;
    },
    routes: [
      _route('/pair', (_) => const PairScreen()),
      _route('/lock', (_) => const LockScreen()),
      StatefulShellRoute.indexedStack(
        builder: (_, _, shell) => AppShell(shell),
        branches: [
          for (final (path, screen) in [
            ('/home', const HomeScreen()),
            ('/sell', const SellScreen()),
            ('/orders', const OrdersScreen()),
          ])
            StatefulShellBranch(routes: [_route(path, (_) => screen)]),
        ],
      ),
      _route('/cart', (_) => const CartScreen()),
      _route('/scan', (_) => const BarcodeScanScreen()),
      _route('/customer', (_) => const CustomerScreen()),
      _route('/order/:key', (s) => OrderDetailScreen(s.pathParameters['key']!)),
      _route('/receipt/:key', (s) => ReceiptScreen(s.pathParameters['key']!)),
      _route('/printer', (_) => const PrinterScreen()),
    ],
  );
});
