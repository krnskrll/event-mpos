import 'package:flutter_test/flutter_test.dart';

import '../../support/fakes.dart';

void main() {
  test('keeps the session across launches but not across reinstalls', () async {
    expect((await fakeTokens(access: 'a', refresh: 'r')).isLoggedIn, isTrue);
    expect((await fakeTokens(access: 'a', refresh: 'r', freshInstall: true)).isLoggedIn, isFalse);
  });
}
