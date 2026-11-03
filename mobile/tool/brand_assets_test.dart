// Renders the launcher icons and launch-screen images from BhoomyLogo.
// Run from mobile/: flutter test tool/brand_assets_test.dart
import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';
import 'dart:ui' as ui;

import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mpos/app/logo.dart';
import 'package:mpos/app/theme.dart';

const _res = 'android/app/src/main/res';
const _assets = 'ios/Runner/Assets.xcassets';

void main() {
  testWidgets('render brand assets', (tester) async {
    await tester.runAsync(
      () => (FontLoader('Lato')..addFont(Future.value(ByteData.sublistView(File('assets/fonts/Lato-Black.ttf').readAsBytesSync())))).load(),
    );

    Widget onShape(ShapeBorder shape, Color color, double size, double mark) => Container(
      width: size,
      height: size,
      decoration: ShapeDecoration(shape: shape, color: color),
      alignment: Alignment.center,
      child: BhoomyLogo(size: mark, tile: false),
    );

    // Android 7.x launcher icon (8+ uses the adaptive icon in mipmap-anydpi-v26).
    await _save(tester, onShape(const CircleBorder(), Brand.forest, 48, 48), {
      for (final (dpi, ratio) in [('mdpi', 1.0), ('hdpi', 1.5), ('xhdpi', 2.0), ('xxhdpi', 3.0), ('xxxhdpi', 4.0)])
        '$_res/mipmap-$dpi/ic_launcher.png': ratio,
    });

    // iOS masks the corners itself and rejects icons with an alpha channel.
    final icons = jsonDecode(File('$_assets/AppIcon.appiconset/Contents.json').readAsStringSync())['images'] as List;
    await _save(tester, onShape(const RoundedRectangleBorder(), Brand.forest, 64, 64), {
      for (final i in icons)
        '$_assets/AppIcon.appiconset/${i['filename']}':
            double.parse((i['size'] as String).split('x').first) * double.parse((i['scale'] as String).replaceAll('x', '')) / 64,
    }, opaque: true);

    // iOS launch screen; Android draws the same circle and mark from vectors.
    await _save(tester, onShape(const CircleBorder(), Brand.forestRaised, 144, 112), {
      for (final (suffix, ratio) in [('', 1.0), ('@2x', 2.0), ('@3x', 3.0)]) '$_assets/LaunchImage.imageset/LaunchImage$suffix.png': ratio,
    });

    final brand = SizedBox(
      width: 200,
      height: 80,
      child: Center(
        child: Text(
          'Bhoomy POS',
          style: AppText.bodyLarge.copyWith(fontWeight: FontWeight.w900, letterSpacing: 0.3, color: Brand.greenPale),
        ),
      ),
    );
    await _save(tester, brand, {
      '$_res/drawable-xxxhdpi/splash_brand.png': 4,
      for (final (suffix, ratio) in [('', 1.0), ('@2x', 2.0), ('@3x', 3.0)]) '$_assets/LaunchBrand.imageset/LaunchBrand$suffix.png': ratio,
    });
  });
}

Future<void> _save(WidgetTester tester, Widget art, Map<String, double> files, {bool opaque = false}) async {
  final key = GlobalKey();
  await tester.pumpWidget(
    Directionality(
      textDirection: TextDirection.ltr,
      child: Center(child: RepaintBoundary(key: key, child: art)),
    ),
  );
  final boundary = key.currentContext!.findRenderObject()! as RenderRepaintBoundary;
  await tester.runAsync(() async {
    for (final MapEntry(key: path, value: ratio) in files.entries) {
      final image = await boundary.toImage(pixelRatio: ratio);
      final bytes = opaque
          ? _rgbPng(image.width, image.height, (await image.toByteData(format: ui.ImageByteFormat.rawRgba))!)
          : (await image.toByteData(format: ui.ImageByteFormat.png))!.buffer.asUint8List();
      File(path)
        ..createSync(recursive: true)
        ..writeAsBytesSync(bytes);
    }
  });
}

/// PNG without an alpha channel, from opaque RGBA pixels.
Uint8List _rgbPng(int w, int h, ByteData rgba) {
  final rows = BytesBuilder();
  for (var y = 0; y < h; y++) {
    rows.addByte(0);
    for (var x = 0; x < w; x++) {
      final i = (y * w + x) * 4;
      rows.add([rgba.getUint8(i), rgba.getUint8(i + 1), rgba.getUint8(i + 2)]);
    }
  }
  Uint8List chunk(String type, List<int> data) {
    final body = [...ascii.encode(type), ...data];
    return (BytesBuilder()
          ..add((ByteData(4)..setUint32(0, data.length)).buffer.asUint8List())
          ..add(body)
          ..add((ByteData(4)..setUint32(0, _crc(body))).buffer.asUint8List()))
        .toBytes();
  }

  final header = ByteData(13)
    ..setUint32(0, w)
    ..setUint32(4, h)
    ..setUint8(8, 8)
    ..setUint8(9, 2);
  return (BytesBuilder()
        ..add([137, 80, 78, 71, 13, 10, 26, 10])
        ..add(chunk('IHDR', header.buffer.asUint8List()))
        ..add(chunk('IDAT', zlib.encode(rows.toBytes())))
        ..add(chunk('IEND', [])))
      .toBytes();
}

int _crc(List<int> bytes) {
  var c = 0xFFFFFFFF;
  for (final b in bytes) {
    c ^= b;
    for (var k = 0; k < 8; k++) {
      c = c & 1 != 0 ? 0xEDB88320 ^ (c >> 1) : c >> 1;
    }
  }
  return c ^ 0xFFFFFFFF;
}
