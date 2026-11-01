import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../app/theme.dart';
import '../../app/widgets.dart';
import '../../core/format/money.dart';
import '../../core/notifications.dart';
import '../../core/printer/printer.dart';
import '../../core/secure/token_storage.dart';
import '../../data/auth/auth_repository.dart';
import '../../data/session/session.dart';
import '../../domain/order.dart';
import '../orders/orders_screen.dart';

class HomeScreen extends ConsumerStatefulWidget {
  const HomeScreen({super.key});

  @override
  ConsumerState<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends ConsumerState<HomeScreen> {
  @override
  void initState() {
    super.initState();
    requestNotificationPermission(); // queued-order updates
  }

  @override
  Widget build(BuildContext context) {
    final event = ref.watch(sessionEventProvider);
    final firstName = (ref.watch(tokenStorageProvider).userName ?? 'Beauty Advisor').split(' ').first;
    final orders = ref.watch(ordersProvider).value ?? const <OrderRecord>[];
    int count(OrderState s) => orders.where((o) => o.state == s).length;
    final revenue = orders.where((o) => o.state == OrderState.confirmed).fold(0.0, (sum, o) => sum + (o.total ?? 0));
    final printer = ref.watch(printerProvider).value;

    return Scaffold(
      backgroundColor: Brand.bg,
      body: Column(
        children: [
          AnnotatedRegion(
            value: SystemUiOverlayStyle.light,
            child: Container(
              color: Brand.navy,
              padding: EdgeInsets.fromLTRB(20, MediaQuery.paddingOf(context).top + 16, 16, 20),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'Hello, $firstName 👋',
                          style: const TextStyle(color: Colors.white, fontSize: 22, fontWeight: FontWeight.bold),
                        ),
                        const SizedBox(height: 4),
                        Text(
                          event == null ? '—' : '${event.name} · ${event.code}',
                          style: const TextStyle(color: Colors.white60, fontSize: 13),
                        ),
                      ],
                    ),
                  ),
                  TextButton(
                    onPressed: () => ref.read(authRepositoryProvider).logout(),
                    style: TextButton.styleFrom(foregroundColor: Colors.white70, backgroundColor: Colors.white12),
                    child: const Text('End session'),
                  ),
                ],
              ),
            ),
          ),
          Expanded(
            child: ListView(
              padding: const EdgeInsets.all(16),
              children: [
                _Tile(
                  icon: Icons.shopping_bag_outlined,
                  title: 'Product Catalogue',
                  subtitle: 'Browse & add items to cart',
                  dark: true,
                  onTap: () => context.go('/catalogue'),
                ),
                const SizedBox(height: 12),
                _Tile(
                  icon: Icons.assignment_outlined,
                  title: 'My Orders Today',
                  subtitle: orders.isEmpty
                      ? 'No orders yet this session.'
                      : [
                          for (final state in OrderState.values)
                            if (count(state) > 0)
                              '${count(state)} ${state == OrderState.pending ? 'syncing' : state.name}',
                        ].join(' · '),
                  trailing: Pill('${orders.length}', color: Brand.navy, background: Brand.navActiveBg),
                  onTap: () => context.go('/orders'),
                ),
                const SizedBox(height: 12),
                Row(
                  children: [
                    Expanded(child: _Stat('Revenue Today', rupiah(revenue), Brand.teal)),
                    const SizedBox(width: 12),
                    Expanded(child: _Stat('Orders', '${orders.length}', Brand.navy)),
                  ],
                ),
                const SizedBox(height: 12),
                _Tile(
                  icon: Icons.print_outlined,
                  title: 'Printer',
                  subtitle: printer?.name ?? 'Checking…',
                  onTap: () => context.push('/printer'),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _Tile extends StatelessWidget {
  const _Tile({
    required this.icon,
    required this.title,
    required this.subtitle,
    required this.onTap,
    this.trailing,
    this.dark = false,
  });

  final IconData icon;
  final String title;
  final String subtitle;
  final VoidCallback onTap;
  final Widget? trailing;
  final bool dark;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: dark ? Brand.navy : Brand.card,
      borderRadius: BorderRadius.circular(16),
      child: ListTile(
        onTap: onTap,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
        contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
        leading: Icon(icon, color: dark ? Colors.white : Brand.ready),
        title: Text(
          title,
          style: TextStyle(color: dark ? Colors.white : Brand.text, fontWeight: FontWeight.bold),
        ),
        subtitle: Text(subtitle, style: TextStyle(color: dark ? Colors.white60 : Brand.textFaint, fontSize: 12)),
        trailing: trailing,
      ),
    );
  }
}

class _Stat extends StatelessWidget {
  const _Stat(this.label, this.value, this.color);

  final String label;
  final String value;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return SectionCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(label, style: const TextStyle(color: Brand.textFaint, fontSize: 11)),
          const SizedBox(height: 4),
          Text(
            value,
            style: TextStyle(color: color, fontSize: 16, fontWeight: FontWeight.bold),
          ),
        ],
      ),
    );
  }
}
