import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../domain/event.dart';

/// The event the BA is selling at. In memory only: a relaunch goes back to event selection.
final sessionEventProvider = NotifierProvider<SessionEvent, AssignedEvent?>(SessionEvent.new);

class SessionEvent extends Notifier<AssignedEvent?> {
  @override
  AssignedEvent? build() => null;

  void start(AssignedEvent event) => state = event;
}
