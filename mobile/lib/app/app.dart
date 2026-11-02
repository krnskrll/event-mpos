import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'router.dart';
import 'theme.dart';

class MposApp extends ConsumerWidget {
  const MposApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return MaterialApp.router(
      title: 'Bhoomy POS',
      theme: appTheme,
      routerConfig: ref.watch(routerProvider),
      debugShowCheckedModeBanner: false,
      builder: (_, child) => AnnotatedRegion(
        value: SystemUiOverlayStyle.dark.copyWith(
          statusBarColor: Colors.transparent,
          systemNavigationBarColor: Brand.canvas,
          systemNavigationBarIconBrightness: Brightness.dark,
        ),
        child: child!,
      ),
    );
  }
}
