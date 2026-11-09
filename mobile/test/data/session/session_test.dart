import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mpos/core/secure/device_storage.dart';
import 'package:mpos/data/session/session.dart';

import '../../support/fakes.dart';

void main() {
  test("checks PINs against Go's bcrypt hashes", () async {
    expect(await checkPin('1234', pin1234), isTrue);
    expect(await checkPin('1235', pin1234), isFalse);
    expect(await checkPin('482913', pin482913), isTrue);
    expect(await checkPin('1234', r'$2b$' + pin1234.substring(4)), isTrue);
    expect(await checkPin('1234', 'not-a-hash'), isFalse);
  });

  test('5 wrong PINs lock that staff member out for a minute', () {
    var now = DateTime(2026, 11, 9, 10);
    final attempts = PinAttempts(() => now);

    expect([for (var i = 0; i < 4; i++) attempts.fail(budi.id)], [4, 3, 2, 1]);
    expect(attempts.waitFor(budi.id), isNull);
    expect(attempts.fail(budi.id), 0);
    expect(attempts.waitFor(budi.id), pinWait);
    expect(attempts.waitFor(sari.id), isNull);

    now = now.add(const Duration(seconds: 59));
    expect(attempts.waitFor(budi.id), const Duration(seconds: 1));
    now = now.add(const Duration(seconds: 1));
    expect(attempts.waitFor(budi.id), isNull);
    expect(attempts.fail(budi.id), 4); // a fresh count after the wait

    attempts.reset(budi.id);
    expect(attempts.fail(budi.id), 4);
  });

  test('removing the pairing signs the staff member out', () async {
    final storage = await fakeDevice();
    final container = ProviderContainer(overrides: [deviceStorageProvider.overrideWithValue(storage)]);
    addTearDown(container.dispose);
    container.listen(signedInProvider, (_, _) {});

    container.read(signedInProvider.notifier).signIn(budi);
    expect(container.read(sessionEventProvider)?.id, 'EV01');
    await storage.clear(removed: true);
    expect((container.read(signedInProvider), container.read(sessionEventProvider)), (null, null));
  });
}
