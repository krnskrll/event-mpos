import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:go_router/go_router.dart';
import 'package:mobile_scanner/mobile_scanner.dart';

import '../../app/theme.dart';

/// Full-screen camera scanner: corner marks around the target, and a bottom panel to type the code
/// or turn on the light. While [onCode] runs (e.g. a result sheet is open) further reads are ignored.
class ScannerView extends StatefulWidget {
  const ScannerView({
    super.key,
    required this.title,
    required this.hint,
    required this.manualLabel,
    required this.onCode,
    this.formats = const [],
    this.actions = const [],
    this.frame = const Size(280, 170),
  });

  final String title;
  final String hint;

  /// Placeholder of the typed-code field, e.g. "Type barcode or SKU".
  final String manualLabel;
  final Future<void> Function(String code) onCode;
  final List<BarcodeFormat> formats;
  final List<Widget> actions;

  /// Wide for barcodes, square for QR codes.
  final Size frame;

  @override
  State<ScannerView> createState() => _ScannerViewState();
}

class _ScannerViewState extends State<ScannerView> {
  late final _controller = MobileScannerController(formats: widget.formats);
  final _typed = TextEditingController();
  bool _busy = false;

  @override
  void dispose() {
    _controller.dispose();
    _typed.dispose();
    super.dispose();
  }

  Future<void> _handle(String? code) async {
    if (_busy || code == null || code.trim().isEmpty) return;
    _busy = true;
    await widget.onCode(code.trim());
    _busy = false;
  }

  Future<void> _submitTyped() async {
    final code = _typed.text;
    FocusScope.of(context).unfocus();
    _typed.clear();
    await _handle(code);
  }

  @override
  Widget build(BuildContext context) {
    return AnnotatedRegion(
      value: SystemUiOverlayStyle.light.copyWith(statusBarColor: Colors.transparent),
      child: Scaffold(
        backgroundColor: Brand.night,
        body: Column(
          children: [
            Expanded(
              child: Stack(
                fit: StackFit.expand,
                children: [
                  MobileScanner(
                    controller: _controller,
                    onDetect: (capture) => _handle(capture.barcodes.firstOrNull?.rawValue),
                    errorBuilder: (_, _) => const SizedBox.shrink(),
                  ),
                  Center(
                    child: Column(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        CustomPaint(size: widget.frame, painter: const _CornersPainter()),
                        const SizedBox(height: 20),
                        Text(
                          widget.hint,
                          textAlign: TextAlign.center,
                          style: AppText.body.copyWith(color: Brand.greenPale),
                        ),
                      ],
                    ),
                  ),
                  SafeArea(
                    child: Padding(
                      padding: const EdgeInsets.fromLTRB(16, 8, 16, 0),
                      child: Row(
                        children: [
                          IconButton(
                            onPressed: context.pop,
                            tooltip: 'Close scanner',
                            icon: const Icon(Icons.close),
                            style: IconButton.styleFrom(
                              backgroundColor: Colors.white.withValues(alpha: 0.14),
                              foregroundColor: Colors.white,
                            ),
                          ),
                          const SizedBox(width: 12),
                          Expanded(
                            child: Text(widget.title, style: AppText.subhead.copyWith(color: Colors.white)),
                          ),
                          ...widget.actions,
                        ],
                      ),
                    ),
                  ),
                ],
              ),
            ),
            // Fallback: type the code, or light up a dim counter.
            Container(
              decoration: const BoxDecoration(
                color: Brand.canvas,
                borderRadius: BorderRadius.vertical(top: Radius.circular(28)),
              ),
              padding: EdgeInsets.fromLTRB(20, 20, 20, 20 + MediaQuery.paddingOf(context).bottom),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text("Code won't scan?", style: AppText.label),
                  const SizedBox(height: 8),
                  Row(
                    children: [
                      Expanded(
                        child: TextField(
                          controller: _typed,
                          textInputAction: TextInputAction.search,
                          onSubmitted: (_) => _submitTyped(),
                          decoration: InputDecoration(
                            hintText: widget.manualLabel,
                            prefixIcon: const Icon(Icons.keyboard_outlined, color: Brand.muted),
                          ),
                        ),
                      ),
                      const SizedBox(width: 10),
                      ValueListenableBuilder(
                        valueListenable: _controller,
                        builder: (_, state, _) {
                          final on = state.torchState == TorchState.on;
                          return IconButton.outlined(
                            onPressed: state.torchState == TorchState.unavailable ? null : _controller.toggleTorch,
                            tooltip: on ? 'Turn off the light' : 'Turn on the light',
                            icon: Icon(on ? Icons.flash_on : Icons.flash_off_outlined),
                            style: IconButton.styleFrom(
                              fixedSize: const Size(52, 52),
                              backgroundColor: on ? Brand.goldSoft : Brand.surface,
                              foregroundColor: on ? Brand.goldInk : Brand.ink,
                              side: const BorderSide(color: Brand.fieldLine),
                              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
                            ),
                          );
                        },
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

/// Rounded corner marks around the scan target, with a scan line through the middle.
class _CornersPainter extends CustomPainter {
  const _CornersPainter();

  @override
  void paint(Canvas canvas, Size size) {
    const arm = 36.0, r = 14.0;
    final corner = Paint()
      ..color = Brand.greenTint
      ..style = PaintingStyle.stroke
      ..strokeWidth = 4
      ..strokeCap = StrokeCap.round;
    final (w, h) = (size.width, size.height);
    for (final (x, y, sx, sy) in [(0.0, 0.0, 1.0, 1.0), (w, 0.0, -1.0, 1.0), (0.0, h, 1.0, -1.0), (w, h, -1.0, -1.0)]) {
      canvas.drawPath(
        Path()
          ..moveTo(x, y + sy * arm)
          ..lineTo(x, y + sy * r)
          ..arcToPoint(Offset(x + sx * r, y), radius: const Radius.circular(r), clockwise: sx * sy > 0)
          ..lineTo(x + sx * arm, y),
        corner,
      );
    }
    canvas.drawLine(
      Offset(16, h / 2),
      Offset(w - 16, h / 2),
      Paint()
        ..color = Brand.greenLight
        ..strokeWidth = 2,
    );
  }

  @override
  bool shouldRepaint(_CornersPainter old) => false;
}
