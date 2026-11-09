import 'package:flutter_test/flutter_test.dart';

import '../../support/fakes.dart';

void main() {
  test('keeps the pairing across launches but not across reinstalls', () async {
    final storage = await fakeDevice();
    expect((storage.isPaired, storage.device!.name, storage.device!.location.code), (true, 'Kasir 1', 'EV01'));
    expect((await fakeDevice(freshInstall: true)).isPaired, isFalse);
  });
}
