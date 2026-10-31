import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:drift/drift.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/db/database.dart';
import '../../core/network/api_client.dart';
import '../../domain/product.dart';

final productRepositoryProvider = Provider(
  (ref) => ProductRepository(ref.watch(dioProvider), ref.watch(databaseProvider)),
);

/// Products come from the local DB, so browsing works offline.
class ProductRepository {
  ProductRepository(this._dio, this._db);

  final Dio _dio;
  final AppDatabase _db;

  Stream<List<Product>> watchProducts(String eventId) {
    final query = _db.select(_db.products)
      ..where((p) => p.eventId.equals(eventId))
      ..orderBy([(p) => OrderingTerm(expression: p.name)]);
    return query.watch().map((rows) => rows.map(_toDomain).toList());
  }

  Future<Product?> product(String eventId, String id) async {
    final query = _db.select(_db.products)..where((p) => p.eventId.equals(eventId) & p.id.equals(id));
    final row = await query.getSingleOrNull();
    return row == null ? null : _toDomain(row);
  }

  /// Pulls live remaining stock. Best effort: offline keeps the last known levels.
  Future<void> refreshStock(String eventId) async {
    final List levels;
    try {
      levels = (await _dio.get<List>('api/me/events/$eventId/stock')).data!;
    } on DioException {
      return;
    }
    await _db.batch((b) {
      for (final level in levels) {
        b.update(
          _db.products,
          ProductsCompanion(remainingStock: Value(level['remaining'] ?? 0)),
          where: (p) => p.eventId.equals(eventId) & p.id.equals(level['productId']),
        );
      }
    });
  }

  static Product _toDomain(ProductRow r) => Product(
    id: r.id,
    sku: r.sku,
    name: r.name,
    brand: r.brand,
    price: r.price,
    imageUrl: r.imageUrl,
    barcodes: r.barcodes.split('\n').where((b) => b.isNotEmpty).toList(),
    remainingStock: r.remainingStock,
    variants: [for (final v in jsonDecode(r.variantsJson)) ProductVariant.fromJson(v)],
  );
}
