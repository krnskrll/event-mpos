import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/secure/token_storage.dart';
import '../../domain/event.dart';

/// The event the BA is selling at. In memory only: a relaunch or sign-out goes back to event selection.
final sessionEventProvider = NotifierProvider<SessionEvent, AssignedEvent?>(SessionEvent.new);

class SessionEvent extends Notifier<AssignedEvent?> {
  @override
  AssignedEvent? build() {
    final tokens = ref.watch(tokenStorageProvider);
    void endOnSignOut() {
      if (!tokens.isLoggedIn) state = null;
    }

    tokens.addListener(endOnSignOut);
    ref.onDispose(() => tokens.removeListener(endOnSignOut));
    return null;
  }

  void start(AssignedEvent event) => state = event;

  void updateStatus(EventStatus status) => state = state?.withStatus(status);
}
