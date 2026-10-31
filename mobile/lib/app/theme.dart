import 'package:flutter/material.dart';

/// Design tokens from the Phase-1 interface design: navy for structure, teal as accent.
abstract final class Brand {
  static const navy = Color(0xFF1F3864);
  static const navyShadow = Color(0x401F3864);
  static const teal = Color(0xFF17A589);
  static const tealSoft = Color(0xFFECFDF5);
  static const tealBorder = Color(0xFFA7F3D0);

  static const bg = Color(0xFFF8FAFC);
  static const card = Color(0xFFFFFFFF);
  static const cardBorder = Color(0xFFF3F4F6);
  static const fieldBg = Color(0xFFFAFAFA);
  static const fieldBorder = Color(0xFFE5E7EB);
  static const navActiveBg = Color(0xFFEEF2FF);

  static const heading = Color(0xFF1F3864);
  static const text = Color(0xFF111827);
  static const textMuted = Color(0xFF6B7280);
  static const textFaint = Color(0xFF9CA3AF);
  static const textGhost = Color(0xFFD1D5DB);

  static const newOrder = Color(0xFF2563EB);
  static const newOrderBg = Color(0xFFEFF6FF);
  static const packing = Color(0xFFD97706);
  static const packingBg = Color(0xFFFFFBEB);
  static const ready = Color(0xFF059669);
  static const readyBg = Color(0xFFECFDF5);
  static const collected = Color(0xFF6B7280);
  static const collectedBg = Color(0xFFF3F4F6);
  static const danger = Color(0xFFEF4444);
  static const warn = Color(0xFFF59E0B);
  static const gold = Color(0xFFD97706);
  static const goldBg = Color(0xFFFEF3C7);
}

final appTheme = ThemeData(
  useMaterial3: true,
  scaffoldBackgroundColor: Brand.bg,
  colorScheme: const ColorScheme.light(
    primary: Brand.navy,
    onPrimary: Colors.white,
    secondary: Brand.teal,
    onSecondary: Colors.white,
    surface: Brand.card,
    onSurface: Brand.text,
    surfaceContainerHighest: Brand.cardBorder,
    onSurfaceVariant: Brand.textMuted,
    error: Brand.danger,
  ),
  inputDecorationTheme: InputDecorationTheme(
    filled: true,
    fillColor: Brand.fieldBg,
    hintStyle: const TextStyle(color: Brand.textFaint),
    enabledBorder: _fieldBorder(Brand.fieldBorder),
    focusedBorder: _fieldBorder(Brand.navy),
  ),
);

OutlineInputBorder _fieldBorder(Color color) => OutlineInputBorder(
  borderRadius: BorderRadius.circular(12),
  borderSide: BorderSide(color: color),
);
