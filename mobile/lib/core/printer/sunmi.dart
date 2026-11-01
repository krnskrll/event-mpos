import 'dart:io';

import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'printer.dart';

/// Bridge to the Sunmi SDK in MainActivity.kt (built-in printer and 2D scanner).
abstract final class Sunmi {
  static const _channel = MethodChannel('mpos/sunmi');
  static bool? _isSunmi;

  static Future<bool> isSunmi() async {
    if (!Platform.isAndroid) return false;
    return _isSunmi ??= await _channel.invokeMethod<bool>('isSunmi') ?? false;
  }
}

/// Codes returned by the Sunmi printer service (updatePrinterState).
PrinterStatus _status(int? code) => switch (code) {
  1 => PrinterStatus.ready,
  4 => PrinterStatus.outOfPaper,
  5 => PrinterStatus.overheated,
  null || 505 => PrinterStatus.unavailable,
  _ => PrinterStatus.error,
};

class SunmiPrinter implements PrinterPort {
  const SunmiPrinter();

  @override
  String get name => 'Sunmi built-in';

  @override
  Future<PrinterStatus> status() async => _status(await Sunmi._channel.invokeMethod<int>('status'));

  @override
  Future<PrinterStatus> print(List<String> lines) async =>
      _status(await Sunmi._channel.invokeMethod<int>('print', {'lines': lines}));
}

/// Codes read by the Sunmi hardware scanner. Never emits on other devices.
final sunmiScansProvider = StreamProvider<String>((ref) async* {
  if (await Sunmi.isSunmi()) {
    yield* const EventChannel('mpos/sunmi/scans').receiveBroadcastStream().cast<String>();
  }
});
