import 'dart:convert';
import 'dart:typed_data';

import 'package:universal_ble/universal_ble.dart';

import 'printer.dart';

/// A 58 mm ESC/POS printer over Bluetooth Low Energy (works on Android and iOS).
class BluetoothPrinter implements PrinterPort {
  const BluetoothPrinter(this.deviceId, this.name);

  final String deviceId;

  @override
  final String name;

  @override
  Future<PrinterStatus> status() async =>
      await UniversalBle.getBluetoothAvailabilityState() == AvailabilityState.poweredOn
      ? PrinterStatus.ready
      : PrinterStatus.unavailable;

  @override
  Future<PrinterStatus> print(List<String> lines) async {
    try {
      await UniversalBle.connect(deviceId, timeout: const Duration(seconds: 10));
      final target = await _writableCharacteristic();
      if (target == null) return PrinterStatus.error;
      final bytes = escPos(lines);
      // Small chunks fit the default BLE packet size on every printer.
      for (var i = 0; i < bytes.length; i += 20) {
        final chunk = bytes.sublist(i, i + 20 > bytes.length ? bytes.length : i + 20);
        await UniversalBle.write(
          deviceId,
          target.service,
          target.characteristic,
          chunk,
          withoutResponse: target.withoutResponse,
        );
      }
      return PrinterStatus.ready;
    } catch (_) {
      return PrinterStatus.unavailable;
    } finally {
      await UniversalBle.disconnect(deviceId).catchError((_) {});
    }
  }

  Future<({String service, String characteristic, bool withoutResponse})?> _writableCharacteristic() async {
    for (final service in await UniversalBle.discoverServices(deviceId)) {
      for (final c in service.characteristics) {
        final withResponse = c.properties.contains(CharacteristicProperty.write);
        if (withResponse || c.properties.contains(CharacteristicProperty.writeWithoutResponse)) {
          return (service: service.uuid, characteristic: c.uuid, withoutResponse: !withResponse);
        }
      }
    }
    return null;
  }
}

/// ESC/POS: initialise, the text, then feed paper past the tear bar.
Uint8List escPos(List<String> lines) =>
    Uint8List.fromList([0x1B, 0x40, ...ascii.encode(_toAscii('${lines.join('\n')}\n')), 0x1B, 0x64, 0x04]);

/// Cheap printers differ in code pages, so print accented letters as their plain form.
String _toAscii(String text) {
  const from = 'àáâãäåèéêëìíîïòóôõöùúûüñçÀÁÂÃÄÅÈÉÊËÌÍÎÏÒÓÔÕÖÙÚÛÜÑÇ—–';
  const to = 'aaaaaaeeeeiiiiooooouuuuncAAAAAAEEEEIIIIOOOOOUUUUNC--';
  return String.fromCharCodes(
    text.runes.map((r) {
      final i = from.indexOf(String.fromCharCode(r));
      return i >= 0 ? to.codeUnitAt(i) : (r < 128 ? r : 0x3F);
    }),
  );
}
