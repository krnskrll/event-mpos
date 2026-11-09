import 'package:flutter/material.dart';

import '../../app/theme.dart';
import '../../app/widgets.dart';
import '../../domain/staff.dart';

/// 3×4 number pad: digits, a blank and backspace. A null [onDigit] disables the pad.
class NumberPad extends StatelessWidget {
  const NumberPad({super.key, required this.onDigit, required this.onBackspace, this.keyColor = Brand.surface});

  final ValueChanged<String>? onDigit;
  final VoidCallback onBackspace;
  final Color keyColor;

  @override
  Widget build(BuildContext context) {
    Widget key(String label) {
      final back = label == '⌫';
      final onTap = back ? onBackspace : (onDigit == null ? null : () => onDigit!(label));
      return Expanded(
        child: Padding(
          padding: const EdgeInsets.all(4),
          child: label.isEmpty
              ? const SizedBox(height: 58)
              : Material(
                  color: back ? Colors.transparent : keyColor,
                  borderRadius: BorderRadius.circular(14),
                  child: InkWell(
                    borderRadius: BorderRadius.circular(14),
                    onTap: onTap,
                    child: SizedBox(
                      height: 58,
                      child: Center(
                        child: back
                            ? const Icon(Icons.backspace_outlined, color: Brand.forest, semanticLabel: 'Delete digit')
                            : Text(
                                label,
                                style: AppText.title.copyWith(
                                  fontWeight: FontWeight.w700,
                                  color: onDigit == null ? Brand.faint : Brand.forest,
                                ),
                              ),
                      ),
                    ),
                  ),
                ),
        ),
      );
    }

    return Column(
      children: [
        for (final row in const [
          ['1', '2', '3'],
          ['4', '5', '6'],
          ['7', '8', '9'],
          ['', '0', '⌫'],
        ])
          Row(children: [for (final label in row) key(label)]),
      ],
    );
  }
}

/// PIN dots: four slots, growing to six as digits are typed.
class PinDots extends StatelessWidget {
  const PinDots(this.length, {super.key, this.error = false});

  final int length;
  final bool error;

  @override
  Widget build(BuildContext context) {
    return Semantics(
      label: 'PIN, $length digits entered',
      child: Row(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          for (var i = 0; i < (length < 4 ? 4 : length); i++)
            Container(
              width: 16,
              height: 16,
              margin: const EdgeInsets.symmetric(horizontal: 7),
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                color: i < length ? Brand.forest : Colors.transparent,
                border: Border.all(color: error && length == 0 ? Brand.errorStrong : Brand.forest, width: 2),
              ),
            ),
        ],
      ),
    );
  }
}

const _avatarColors = [Brand.green, Color(0xFFB07A16), Brand.info, Color(0xFF7A4E9C)];

/// Initials on a colour that stays the same for each staff member.
class StaffAvatar extends StatelessWidget {
  const StaffAvatar(this.staff, {super.key, this.radius = 26});

  final StaffMember staff;
  final double radius;

  @override
  Widget build(BuildContext context) => CircleAvatar(
    radius: radius,
    backgroundColor: _avatarColors[staff.id.codeUnits.fold(0, (a, b) => a + b) % _avatarColors.length],
    child: Text(
      initials(staff.name),
      style: AppText.label.copyWith(color: Colors.white, fontWeight: FontWeight.w900, fontSize: radius * 0.68),
    ),
  );
}

/// "0:42" for a lockout countdown.
String countdown(Duration d) => '${d.inMinutes}:${(d.inSeconds % 60).toString().padLeft(2, '0')}';
