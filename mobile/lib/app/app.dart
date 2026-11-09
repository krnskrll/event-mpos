import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../data/device/device_repository.dart';
import '../data/session/session.dart';
import 'router.dart';
import 'theme.dart';

class MposApp extends ConsumerWidget {
  const MposApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    ref.watch(deviceSyncProvider);
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
        child: IdleLock(child: child!),
      ),
    );
  }
}

/// Locks the device after [idleLock] without a touch anywhere in the app, also across a trip to
/// the background.
class IdleLock extends ConsumerStatefulWidget {
  const IdleLock({super.key, required this.child});

  final Widget child;

  @override
  ConsumerState<IdleLock> createState() => _IdleLockState();
}

class _IdleLockState extends ConsumerState<IdleLock> {
  Timer? _timer;
  DateTime _last = DateTime.now();
  late final AppLifecycleListener _lifecycle;

  @override
  void initState() {
    super.initState();
    _lifecycle = AppLifecycleListener(onResume: _resume);
  }

  @override
  void dispose() {
    _timer?.cancel();
    _lifecycle.dispose();
    super.dispose();
  }

  void _touch([_]) {
    _last = DateTime.now();
    _timer?.cancel();
    _timer = ref.read(signedInProvider) == null ? null : Timer(idleLock, _lock);
  }

  void _resume() => DateTime.now().difference(_last) >= idleLock ? _lock() : _touch();

  void _lock() => ref.read(signedInProvider.notifier).lock();

  @override
  Widget build(BuildContext context) {
    ref.listen(signedInProvider, (_, _) => _touch());
    return Listener(behavior: HitTestBehavior.translucent, onPointerDown: _touch, child: widget.child);
  }
}
