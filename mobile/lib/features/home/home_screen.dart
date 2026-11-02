import 'package:flutter/material.dart';
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
    final firstName = (ref.watch(tokenStorageProvider).userName ?? 'there').split(' ').first;
    final orders = ref.watch(ordersProvider).value ?? const <OrderRecord>[];
    int count(OrderState s) => orders.where((o) => o.state == s).length;
    final revenue = orders.where((o) => o.state == OrderState.confirmed).fold(0.0, (sum, o) => sum + (o.total ?? 0));
    final printer = ref.watch(printerProvider).value;

    return Scaffold(
      body: Column(
        children: [
          ScreenHeader(
            title: 'Hi, $firstName',
            subtitle: event == null ? null : '${event.name} · ${event.code}',
            actions: [
              TextButton(onPressed: () => ref.read(authRepositoryProvider).logout(), child: const Text('End session')),
            ],
          ),
          Expanded(
            child: ListView(
              padding: const EdgeInsets.fromLTRB(16, 8, 16, 16),
              children: [
                _Tile(
                  icon: Icons.shopping_bag_outlined,
                  title: 'Product Catalogue',
                  subtitle: 'Browse & add items to cart',
                  primary: true,
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
                              '${count(state)} ${state == OrderState.pending ? 'sending' : state.name}',
                        ].join(' · '),
                  trailing: Pill('${orders.length}', color: Brand.forest, background: Brand.greenSoft),
                  onTap: () => context.go('/orders'),
                ),
                const SizedBox(height: 12),
                Row(
                  children: [
                    Expanded(child: _Stat('Revenue today', rupiah(revenue))),
                    const SizedBox(width: 12),
                    Expanded(child: _Stat('Orders', '${orders.length}')),
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
    this.primary = false,
  });

  final IconData icon;
  final String title;
  final String subtitle;
  final VoidCallback onTap;
  final Widget? trailing;
  final bool primary;

  @override
  Widget build(BuildContext context) {
    final shape = RoundedRectangleBorder(
      borderRadius: BorderRadius.circular(16),
      side: primary ? BorderSide.none : const BorderSide(color: Brand.cardLine),
    );
    return Material(
      color: primary ? Brand.forest : Brand.surface,
      shape: shape,
      child: ListTile(
        onTap: onTap,
        shape: shape,
        contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
        leading: Icon(icon, color: primary ? Brand.gold : Brand.green),
        title: Text(title, style: AppText.bodyLarge.copyWith(color: primary ? Colors.white : Brand.forest)),
        subtitle: Text(subtitle, style: AppText.caption.copyWith(color: primary ? Brand.greenTint : Brand.muted)),
        trailing: trailing,
      ),
    );
  }
}

class _Stat extends StatelessWidget {
  const _Stat(this.label, this.value);

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return SectionCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(label, style: AppText.caption),
          const SizedBox(height: 4),
          Text(value, style: AppText.heading.copyWith(fontWeight: FontWeight.w900)),
        ],
      ),
    );
  }
}
