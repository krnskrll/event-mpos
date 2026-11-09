import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:mpos/core/secure/device_storage.dart';
import 'package:mpos/domain/device.dart';
import 'package:mpos/domain/event.dart';
import 'package:mpos/domain/staff.dart';
import 'package:shared_preferences/shared_preferences.dart';

const testLocation = AssignedEvent(id: 'EV01', code: 'EV01', name: 'Roadshow GI', status: EventStatus.active);
const testDevice = PairedDevice(id: 'D1', name: 'Kasir 1', storeName: 'RBeauty Shops', location: testLocation);

/// Go's bcrypt (cost 10) of "1234" and "482913", as `GET /api/device/staff` returns them.
const pin1234 = r'$2a$10$1fBeqc51tNzAXxUAHhVFyOkbpObzonzNnl56aGyMy6LYjz5DHFX8m';
const pin482913 = r'$2a$10$mNSJl13emdsHoPiqfhGPLeSrLFKPqScfH0ePfKEsjcLGaoO.25ufq';

const budi = StaffMember(id: 's-budi', name: 'Budi Santoso', isSupervisor: false, pinHash: pin1234);
const sari = StaffMember(id: 's-sari', name: 'Sari Wulandari', isSupervisor: true, pinHash: pin482913);

/// Device storage backed by in-memory platform mocks; paired to [testDevice] by default.
Future<DeviceStorage> fakeDevice({bool paired = true, bool freshInstall = false}) async {
  SharedPreferences.setMockInitialValues({if (!freshInstall) 'installed': true});
  FlutterSecureStorage.setMockInitialValues({
    if (paired) 'device_token': 't1',
    if (paired) 'device': jsonEncode(testDevice.toJson()),
  });
  return DeviceStorage.open();
}

Future<SharedPreferences> fakePrefs() async {
  SharedPreferences.setMockInitialValues({});
  return SharedPreferences.getInstance();
}

/// Answers like the backend: [routes] are canned `(status, body)` replies by path; any other call
/// needs the device token [validToken].
class FakeServer implements HttpClientAdapter {
  String validToken = 't1';
  final routes = <String, (int, Object)>{};
  final requests = <RequestOptions>[];
  Duration delay = Duration.zero;

  @override
  Future<ResponseBody> fetch(RequestOptions options, Stream<Uint8List>? _, Future<void>? _) async {
    requests.add(options);
    await Future.delayed(delay);
    if (routes[options.path] case (final status, final body)) return _json(status, body);
    return _json(options.headers['Authorization'] == 'Device $validToken' ? 200 : 401, {});
  }

  @override
  void close({bool force = false}) {}

  static ResponseBody _json(int status, Object body) => ResponseBody.fromString(
    jsonEncode(body),
    status,
    headers: {
      Headers.contentTypeHeader: [Headers.jsonContentType],
    },
  );
}

/// Replies of a freshly paired device at an event, as the server sends them.
void serveDevice(FakeServer server) => server.routes.addAll({
  'api/device/pair': (
    201,
    {
      'token': 't1',
      'device': {'id': 'D3', 'name': 'Kasir 3'},
      'location': {'id': 'EV01', 'name': 'Roadshow GI', 'type': 'event', 'status': 'ACTIVE'},
      'storeName': 'RBeauty Shops',
    },
  ),
  'api/device': (
    200,
    {
      'device': {'id': 'D3', 'name': 'Kasir 3'},
      'location': {'id': 'EV01', 'name': 'Roadshow GI', 'type': 'event', 'status': 'ACTIVE'},
      'storeName': 'RBeauty Shops',
    },
  ),
  'api/me/event': (
    200,
    {'id': 'EV01', 'code': 'RGI', 'name': 'Roadshow GI', 'type': 'event', 'status': 'ACTIVE', 'endDate': '2027-03-20'},
  ),
  'api/device/staff': (
    200,
    [
      {'id': budi.id, 'name': budi.name, 'role': 'KASIR', 'pinHash': pin1234},
      {'id': sari.id, 'name': sari.name, 'role': 'SUPERVISOR', 'pinHash': pin482913},
    ],
  ),
  'api/me/events/EV01/snapshot/hash': (200, {'snapshotHash': 'v1'}),
  'api/me/events/EV01/snapshot': (200, {'version': 'v1', 'products': [], 'members': [], 'promotions': []}),
  'api/me/events/EV01/stock': (200, []),
  'api/me/events/EV01/number-blocks': (201, {'businessDate': '20261109', 'rangeStart': 1, 'rangeEnd': 50}),
});
