import 'dart:isolate';

import 'package:bcrypt/bcrypt.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/secure/device_storage.dart';
import '../../domain/device.dart';
import '../../domain/event.dart';
import '../../domain/staff.dart';

/// The paired device; rebuilt whenever the pairing changes.
final deviceProvider = Provider<PairedDevice?>((ref) {
  final storage = ref.watch(deviceStorageProvider);
  storage.addListener(ref.invalidateSelf);
  ref.onDispose(() => storage.removeListener(ref.invalidateSelf));
  return storage.device;
});

/// Where this device sells: the one location it is paired to.
final sessionEventProvider = Provider<AssignedEvent?>((ref) => ref.watch(deviceProvider)?.location);

/// The staff member signed in on this device; null while the lock screen shows. In memory only, so a
/// relaunch starts locked. A new or removed pairing signs out.
final signedInProvider = NotifierProvider<SignedIn, StaffMember?>(SignedIn.new);

class SignedIn extends Notifier<StaffMember?> {
  @override
  StaffMember? build() {
    ref.watch(deviceProvider.select((d) => d?.id));
    return null;
  }

  void signIn(StaffMember staff) => state = staff;

  void lock() => state = null;
}

/// Device lock rules: 5 wrong PINs make that staff member wait a minute; 5 idle minutes lock the device.
const maxPinTries = 5;
const pinWait = Duration(minutes: 1);
const idleLock = Duration(minutes: 5);

final pinAttemptsProvider = Provider((_) => PinAttempts());

/// Wrong PIN tries per staff member, in memory.
class PinAttempts {
  PinAttempts([this._now = DateTime.now]);

  final DateTime Function() _now;
  final _fails = <String, int>{};
  final _until = <String, DateTime>{};

  /// Time left before [staffId] may try again; null when not locked out.
  Duration? waitFor(String staffId) {
    final left = _until[staffId]?.difference(_now());
    if (left == null || left <= Duration.zero) {
      _until.remove(staffId);
      return null;
    }
    return left;
  }

  /// Records a wrong PIN; returns the tries left (0 = locked out for [pinWait]).
  int fail(String staffId) {
    final fails = (_fails[staffId] ?? 0) + 1;
    if (fails < maxPinTries) return maxPinTries - (_fails[staffId] = fails);
    _fails.remove(staffId);
    _until[staffId] = _now().add(pinWait);
    return 0;
  }

  void reset(String staffId) => _fails.remove(staffId);
}

/// Checks a PIN against its bcrypt hash off the UI thread (cost 10 takes a moment in pure Dart).
Future<bool> checkPin(String pin, String hash) => Isolate.run(() {
  try {
    return BCrypt.checkpw(pin, hash);
  } catch (_) {
    return false; // malformed hash
  }
});

final pinCheckProvider = Provider<Future<bool> Function(String pin, String hash)>((_) => checkPin);
