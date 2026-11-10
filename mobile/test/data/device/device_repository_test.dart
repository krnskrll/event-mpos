import 'package:dio/dio.dart';
import 'package:drift/native.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mpos/core/db/database.dart';
import 'package:mpos/core/network/api_failure.dart';
import 'package:mpos/core/network/device_interceptor.dart';
import 'package:mpos/core/secure/device_storage.dart';
import 'package:mpos/data/device/device_repository.dart';
import 'package:mpos/data/event/event_repository.dart';

import '../../support/fakes.dart';

void main() {
  late AppDatabase db;
  late FakeServer server;
  late DeviceStorage storage;
  late DeviceRepository repo;

  setUp(() async {
    db = AppDatabase(NativeDatabase.memory());
    server = FakeServer();
    serveDevice(server);
    storage = await fakeDevice(paired: false);
    final dio = Dio()
      ..httpClientAdapter = server
      ..interceptors.add(DeviceInterceptor(storage, () => null));
    repo = DeviceRepository(dio, db, storage, EventRepository(dio, db));
  });
  tearDown(() => db.close());

  test('pairing stores the token, drops old number blocks, then downloads staff and snapshot', () async {
    await db
        .into(db.numberBlocks)
        .insert(
          NumberBlocksCompanion.insert(
            eventId: 'EV01',
            businessDate: '20261108',
            rangeStart: 1,
            rangeEnd: 50,
            nextUnused: 3,
          ),
        );

    await repo.pair('482913');
    expect(server.requests.first.data['code'], '482913');
    expect(server.requests.first.headers.containsKey('Authorization'), isFalse);
    expect((storage.token, storage.device!.name, storage.device!.location.name), ('t1', 'Kasir 3', 'Roadshow GI'));
    expect(await db.select(db.numberBlocks).get(), isEmpty);

    final counts = await repo.prepare();
    expect((counts.products, counts.staff), (0, 2));
    expect(storage.device!.location.code, 'RGI'); // the pairing reply has no code; /me/event fills it
    expect(server.requests.last.headers['Authorization'], 'Device t1');
    final staff = await repo.staff();
    expect([for (final s in staff) (s.name, s.isSupervisor)], [(budi.name, false), (sari.name, true)]);
    expect(staff.first.pinHash, pin1234);
  });

  test('pairing errors read plainly', () async {
    for (final (status, message) in [
      (404, 'Wrong code, or older than 10 minutes. Ask for a new code in Bhoomy.'),
      (409, 'This location is closed. Devices can no longer be added to it.'),
      (429, 'Too many tries. Wait a minute, then try again.'),
    ]) {
      server.routes['api/device/pair'] = (status, {'message': 'server text'});
      await expectLater(repo.pair('000000'), throwsA(isA<ApiFailure>().having((e) => e.message, 'message', message)));
    }
    expect(storage.isPaired, isFalse);
  });

  test('a staff refresh replaces the cached list', () async {
    await repo.refreshStaff();
    server.routes['api/device/staff'] = (
      200,
      [
        {'id': sari.id, 'name': sari.name, 'role': 'SUPERVISOR', 'pinHash': pin482913},
      ],
    );
    await repo.refreshStaff();
    expect((await repo.staff()).map((s) => s.id), [sari.id]);
  });
}
