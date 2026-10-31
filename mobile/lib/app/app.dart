import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'offline_banner.dart';
import 'router.dart';
import 'theme.dart';

class MposApp extends ConsumerWidget {
  const MposApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return MaterialApp.router(
      title: 'event-mpos',
      theme: appTheme,
      routerConfig: ref.watch(routerProvider),
      debugShowCheckedModeBanner: false,
      builder: (_, child) => OfflineBanner(child: child!),
    );
  }
}
