import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../core/connectivity.dart';
import 'theme.dart';

/// Wraps every screen: slides in a warning bar while offline. Selling continues offline;
/// only uploads wait for the network.
class OfflineBanner extends ConsumerWidget {
  const OfflineBanner({super.key, required this.child});

  final Widget child;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final online = ref.isOnline;
    return Column(
      children: [
        AnimatedSize(
          duration: const Duration(milliseconds: 250),
          child: online
              ? const SizedBox(width: double.infinity)
              : Material(
                  color: Brand.warn,
                  child: Padding(
                    padding: EdgeInsets.fromLTRB(16, MediaQuery.paddingOf(context).top + 8, 16, 8),
                    child: const Row(
                      children: [
                        Icon(Icons.cloud_off_outlined, color: Colors.white, size: 16),
                        SizedBox(width: 8),
                        Expanded(
                          child: Text(
                            'Offline — keep selling; orders sync when reconnected',
                            style: TextStyle(color: Colors.white, fontSize: 12, fontWeight: FontWeight.w600),
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
        ),
        // The banner already covers the status bar, so screens below drop their top inset.
        Expanded(
          child: MediaQuery.removePadding(context: context, removeTop: !online, child: child),
        ),
      ],
    );
  }
}
