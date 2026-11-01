import 'package:drift/native.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mpos/core/db/database.dart';
import 'package:mpos/core/secure/token_storage.dart';
import 'package:mpos/data/cart/cart_notifier.dart';
import 'package:mpos/data/session/session.dart';
import 'package:mpos/domain/event.dart';

import '../../support/fakes.dart';

const _event = AssignedEvent(id: 'EV01', code: 'EV01', name: 'Roadshow', status: EventStatus.active);

void main() {
  late AppDatabase db;

  setUp(() => db = AppDatabase(NativeDatabase.memory()));
  tearDown(() => db.close());

  Future<ProviderContainer> session() async {
    final container = ProviderContainer(
      overrides: [
        databaseProvider.overrideWithValue(db),
        tokenStorageProvider.overrideWithValue(await fakeTokens(access: 'a', refresh: 'r')),
      ],
    );
    container.read(sessionEventProvider.notifier).start(_event);
    container.read(cartProvider);
    return container;
  }

  test('the live cart survives a restart; a draft parks and resumes the cart', () async {
    final first = await session();
    first.read(cartProvider.notifier)
      ..add('p1', qty: 2)
      ..add('p2', variantId: 'v1');
    await Future.delayed(const Duration(milliseconds: 400)); // autosave debounce
    first.dispose();

    final restarted = await session();
    await Future.delayed(const Duration(milliseconds: 50));
    expect(restarted.read(cartProvider).items, {'p1|': 2, 'p2|v1': 1});

    await restarted.read(cartProvider.notifier).saveDraft('Aisha');
    expect(restarted.read(cartProvider).items, isEmpty);
    final draft = (await db.select(db.cartDrafts).get()).singleWhere((d) => !d.live);
    await restarted.read(cartProvider.notifier).resumeDraft(draft.id);
    expect(restarted.read(cartProvider).count, 3);
    restarted.dispose();
  });
}
