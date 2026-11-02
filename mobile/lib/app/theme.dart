import 'package:flutter/material.dart';

/// Bhoomy Brand Guidelines v2.0 tokens, with softer corners for touch.
abstract final class Brand {
  static const green = Color(0xFF2E7D53);
  static const greenPressed = Color(0xFF22633F);
  static const greenSoft = Color(0xFFEDF6F0);
  static const greenTint = Color(0xFFA9D6BC);
  static const greenPale = Color(0xFFD3EBDE); // body text on forest
  static const forest = Color(0xFF143A24);
  static const forestRaised = Color(0xFF1A4E30); // buttons on the forest rail
  static const greenLight = Color(0xFF54A87A); // status dot on forest
  static const gold = Color(0xFFD99A2B);
  static const goldSoft = Color(0xFFFBF3E3);
  static const goldInk = Color(0xFF6D4B00);
  static const goldTint = Color(0xFFF6E4BE);
  static const onGold = Color(0xFF2B1D00); // text on gold

  static const canvas = Color(0xFFF7F6F1);
  static const surface = Color(0xFFFFFFFF);
  static const sunken = Color(0xFFEDF1F6);
  static const divider = Color(0xFFDCE3EC);
  static const cardLine = Color(0xFFE7E4DA);
  static const fieldLine = Color(0xFF8A94A0);

  static const ink = Color(0xFF33415A);
  static const slate = Color(0xFF47566B); // secondary text on the canvas
  static const muted = Color(0xFF64748B);
  static const faint = Color(0xFF94A3B4);

  // Status: the base colour for dots and fills, the ink shade for text on white or on the soft background.
  static const success = Color(0xFF12885E);
  static const successInk = Color(0xFF0B5239);
  static const successSoft = Color(0xFFE7F6EF);
  static const warning = Color(0xFFC98A00);
  static const warningInk = goldInk;
  static const warningSoft = goldSoft;
  static const error = Color(0xFFD64550);
  static const errorInk = Color(0xFF8A242C);
  static const errorStrong = Color(0xFFB22F39); // white text on it passes contrast
  static const errorSoft = Color(0xFFFCECEE);
  static const info = Color(0xFF2F6FB5);
  static const infoInk = Color(0xFF1C436E);
  static const infoSoft = Color(0xFFEAF2FA);
}

/// Soft, green-tinted elevation: cards rest, panels float.
abstract final class AppShadows {
  static const card = [BoxShadow(color: Color(0x14143A24), blurRadius: 4, offset: Offset(0, 4), spreadRadius: -2)];
  static const panel = [BoxShadow(color: Color(0x1F143A24), blurRadius: 16, offset: Offset(0, 8), spreadRadius: -4)];
  static const paper = [BoxShadow(color: Color(0x400F1829), blurRadius: 18, offset: Offset(0, 8), spreadRadius: -8)];
}

/// The type scale: Lato, tabular figures so prices line up. Each style names the family because
/// button and chip text don't inherit the app font.
abstract final class AppText {
  static const _numbers = [FontFeature.tabularFigures()];
  static const _lato = 'Lato';
  static const total = TextStyle(
    fontFamily: _lato,
    fontSize: 28,
    height: 1.2,
    fontWeight: FontWeight.w900,
    color: Brand.forest,
    fontFeatures: _numbers,
  );
  static const title = TextStyle(
    fontFamily: _lato,
    fontSize: 24,
    height: 1.25,
    fontWeight: FontWeight.w900,
    color: Brand.forest,
  );
  static const heading = TextStyle(
    fontFamily: _lato,
    fontSize: 20,
    height: 1.3,
    fontWeight: FontWeight.w700,
    color: Brand.forest,
  );
  static const subhead = TextStyle(
    fontFamily: _lato,
    fontSize: 18,
    height: 1.3,
    fontWeight: FontWeight.w700,
    color: Brand.forest,
  );
  static const bodyLarge = TextStyle(
    fontFamily: _lato,
    fontSize: 16,
    height: 1.4,
    fontWeight: FontWeight.w700,
    color: Brand.ink,
    fontFeatures: _numbers,
  );
  static const body = TextStyle(fontFamily: _lato, fontSize: 14, height: 1.4, color: Brand.ink, fontFeatures: _numbers);
  static const label = TextStyle(
    fontFamily: _lato,
    fontSize: 13,
    height: 1.4,
    fontWeight: FontWeight.w700,
    color: Brand.ink,
  );
  static const button = TextStyle(fontFamily: _lato, fontSize: 15, fontWeight: FontWeight.w700);
  static const caption = TextStyle(
    fontFamily: _lato,
    fontSize: 12,
    height: 1.35,
    color: Brand.muted,
    fontFeatures: _numbers,
  );
}

RoundedRectangleBorder _rounded(double radius, [BorderSide side = BorderSide.none]) =>
    RoundedRectangleBorder(borderRadius: BorderRadius.circular(radius), side: side);

OutlineInputBorder _field(Color color, [double width = 1]) => OutlineInputBorder(
  borderRadius: BorderRadius.circular(12),
  borderSide: BorderSide(color: color, width: width),
);

WidgetStateProperty<T> _selected<T>(T on, T off) =>
    WidgetStateProperty.resolveWith((s) => s.contains(WidgetState.selected) ? on : off);

final appTheme = ThemeData(
  useMaterial3: true,
  fontFamily: 'Lato',
  scaffoldBackgroundColor: Brand.canvas,
  colorScheme: const ColorScheme.light(
    primary: Brand.green,
    onPrimary: Colors.white,
    secondary: Brand.forest,
    onSecondary: Colors.white,
    secondaryContainer: Brand.greenSoft, // tonal buttons, selected tab
    onSecondaryContainer: Brand.forest,
    surface: Brand.surface,
    onSurface: Brand.ink,
    surfaceContainerHighest: Brand.sunken,
    onSurfaceVariant: Brand.muted,
    outline: Brand.fieldLine,
    outlineVariant: Brand.divider,
    error: Brand.error,
  ),
  textTheme: const TextTheme(bodyMedium: AppText.body, bodyLarge: AppText.body, titleMedium: AppText.subhead),
  filledButtonTheme: FilledButtonThemeData(
    style: FilledButton.styleFrom(minimumSize: const Size(64, 48), shape: _rounded(12), textStyle: AppText.button),
  ),
  outlinedButtonTheme: OutlinedButtonThemeData(
    style: OutlinedButton.styleFrom(
      minimumSize: const Size(64, 48),
      foregroundColor: Brand.ink,
      side: const BorderSide(color: Brand.fieldLine),
      shape: _rounded(12),
      textStyle: AppText.button,
    ),
  ),
  textButtonTheme: TextButtonThemeData(
    style: TextButton.styleFrom(
      foregroundColor: Brand.green,
      minimumSize: const Size(44, 44),
      textStyle: AppText.button.copyWith(fontSize: 14),
    ),
  ),
  chipTheme: ChipThemeData(
    backgroundColor: Brand.surface,
    selectedColor: Brand.green,
    side: WidgetStateBorderSide.resolveWith(
      (s) => BorderSide(color: s.contains(WidgetState.selected) ? Brand.green : Brand.divider),
    ),
    shape: const StadiumBorder(),
    showCheckmark: false,
    labelStyle: AppText.label.copyWith(
      color: WidgetStateColor.resolveWith((s) => s.contains(WidgetState.selected) ? Colors.white : Brand.ink),
    ),
    padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 6),
  ),
  inputDecorationTheme: InputDecorationTheme(
    filled: true,
    fillColor: Brand.surface,
    hintStyle: const TextStyle(color: Brand.faint),
    contentPadding: const EdgeInsets.symmetric(horizontal: 14, vertical: 14),
    enabledBorder: _field(Brand.fieldLine),
    focusedBorder: _field(Brand.green, 2),
    errorBorder: _field(Brand.error),
    focusedErrorBorder: _field(Brand.error, 2),
  ),
  cardTheme: CardThemeData(
    color: Brand.surface,
    elevation: 0,
    margin: EdgeInsets.zero,
    shape: _rounded(16, const BorderSide(color: Brand.cardLine)),
  ),
  appBarTheme: const AppBarTheme(
    backgroundColor: Brand.canvas,
    foregroundColor: Brand.forest,
    elevation: 0,
    scrolledUnderElevation: 0,
    titleTextStyle: AppText.subhead,
  ),
  navigationBarTheme: NavigationBarThemeData(
    backgroundColor: Brand.surface,
    indicatorColor: Brand.greenSoft,
    surfaceTintColor: Colors.transparent,
    iconTheme: _selected(const IconThemeData(color: Brand.forest), const IconThemeData(color: Brand.muted)),
    labelTextStyle: _selected(
      const TextStyle(fontSize: 12, fontWeight: FontWeight.w700, color: Brand.forest),
      const TextStyle(fontSize: 12, fontWeight: FontWeight.w700, color: Brand.muted),
    ),
  ),
  navigationRailTheme: const NavigationRailThemeData(
    backgroundColor: Brand.forest,
    indicatorColor: Brand.green,
    selectedIconTheme: IconThemeData(color: Colors.white),
    unselectedIconTheme: IconThemeData(color: Brand.greenTint),
    selectedLabelTextStyle: TextStyle(fontSize: 12, fontWeight: FontWeight.w700, color: Colors.white),
    unselectedLabelTextStyle: TextStyle(fontSize: 12, fontWeight: FontWeight.w700, color: Brand.greenTint),
  ),
  badgeTheme: const BadgeThemeData(backgroundColor: Brand.forest, textColor: Colors.white),
  dialogTheme: DialogThemeData(backgroundColor: Brand.surface, shape: _rounded(24), titleTextStyle: AppText.heading),
  bottomSheetTheme: const BottomSheetThemeData(
    backgroundColor: Brand.surface,
    showDragHandle: true,
    shape: RoundedRectangleBorder(borderRadius: BorderRadius.vertical(top: Radius.circular(24))),
  ),
  snackBarTheme: SnackBarThemeData(
    behavior: SnackBarBehavior.floating,
    backgroundColor: Brand.forest,
    shape: _rounded(12),
  ),
  progressIndicatorTheme: const ProgressIndicatorThemeData(color: Brand.green, linearTrackColor: Brand.sunken),
  dividerTheme: const DividerThemeData(color: Brand.divider, space: 1),
);
