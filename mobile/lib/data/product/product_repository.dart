import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:drift/drift.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/db/database.dart';
import '../../core/network/api_client.dart';
import '../../domain/cart.dart';
import '../../domain/product.dart';
import '../session/session.dart';

final productRepositoryProvider = Provider(
  (ref) => ProductRepository(ref.watch(dioProvider), ref.watch(databaseProvider)),
);

/// Live products of the selected event.
final productsProvider = StreamProvider.autoDispose(
  (ref) => ref.watch(productRepositoryProvider).watchProducts(ref.watch(sessionEventProvider)!.id),
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

  /// A scanned or typed code: barcode first, then SKU, then part of the name.
  Future<Product?> findByCode(String eventId, String code) async {
    final products = await watchProducts(eventId).first;
    final c = code.trim().toLowerCase();
    return products.where((p) => p.barcodes.contains(code.trim())).firstOrNull ??
        products.where((p) => p.sku.toLowerCase() == c).firstOrNull ??
        products.where((p) => c.isNotEmpty && p.name.toLowerCase().contains(c)).firstOrNull;
  }

  /// Local best guess after a sale (negative) or a cancel (positive) until the next read.
  Future<void> adjustStock(String eventId, String productId, String? variantId, int delta) async {
    final row = await (_db.select(
      _db.products,
    )..where((p) => p.eventId.equals(eventId) & p.id.equals(productId))).getSingleOrNull();
    if (row == null || !row.stockTracked) return;
    if (variantId == null) {
      await (_db.update(_db.products)..where((p) => p.eventId.equals(eventId) & p.id.equals(productId))).write(
        ProductsCompanion(remainingStock: Value(row.remainingStock + delta)),
      );
      return;
    }
    final stock = Map<String, dynamic>.from(jsonDecode(row.variantStockJson));
    stock[variantId] = ((stock[variantId] as int?) ?? 0) + delta;
    await (_db.update(_db.products)..where((p) => p.eventId.equals(eventId) & p.id.equals(productId))).write(
      ProductsCompanion(variantStockJson: Value(jsonEncode(stock))),
    );
  }

  /// The place this location sells from (null = untracked), as of the last stock read.
  Future<String?> stockPlace(String eventId) async =>
      (await (_db.select(_db.eventStocks)..where((s) => s.eventId.equals(eventId))).getSingleOrNull())?.placeId;

  /// Reads what this location may sell from Bhoomy (6c): per product or variant, minus this
  /// device's sales Bhoomy may not have yet. Best effort: offline keeps the last numbers.
  Future<void> refreshStock(String eventId) async {
    final Map<String, dynamic> body;
    try {
      body = (await _dio.get<Map<String, dynamic>>('api/me/events/$eventId/stock-levels')).data!;
    } on DioException {
      return;
    }
    final tracked = body['tracked'] == true;
    final unsent = await _unsent(eventId);
    final product = <String, int>{};
    final variants = <String, Map<String, int>>{};
    for (final l in (body['levels'] as List? ?? [])) {
      final p = l['productId'] as String, v = l['variantId'] as String?;
      final left = (l['available'] as int? ?? 0) - (unsent[cartKey(p, v)] ?? 0);
      if (v == null) {
        product[p] = left;
      } else {
        (variants[p] ??= {})[v] = left;
      }
    }
    final rows = await (_db.select(_db.products)..where((p) => p.eventId.equals(eventId))).get();
    await _db.batch((b) {
      b.insert(
        _db.eventStocks,
        EventStocksCompanion.insert(
          eventId: eventId,
          tracked: tracked,
          placeId: Value(body['placeId'] as String?),
          refreshedAt: DateTime.now(),
        ),
        mode: InsertMode.insertOrReplace,
      );
      for (final r in rows) {
        b.update(
          _db.products,
          ProductsCompanion(
            stockTracked: Value(tracked),
            remainingStock: Value(product[r.id] ?? 0),
            variantStockJson: Value(jsonEncode(variants[r.id] ?? {})),
          ),
          where: (p) => p.eventId.equals(eventId) & p.id.equals(r.id),
        );
      }
    });
  }

  /// Quantities sold on this device that Bhoomy may not count yet: not uploaded, or uploaded
  /// within the last minute (POS sends every 15 s, the stock read is up to 30 s old).
  Future<Map<String, int>> _unsent(String eventId) async {
    final since = DateTime.now().subtract(const Duration(minutes: 1));
    final orders =
        await (_db.select(_db.pendingOrders)..where(
              (o) =>
                  o.eventId.equals(eventId) &
                  o.status.isNotIn(['CANCELLED']) &
                  (o.status.isNotIn(['CONFIRMED']) | o.createdAt.isBiggerThanValue(since)),
            ))
            .get();
    final out = <String, int>{};
    for (final o in orders) {
      for (final l in jsonDecode(o.linesJson) as List) {
        final k = cartKey(l['productId'] as String, l['variantId'] as String?);
        out[k] = (out[k] ?? 0) + (l['qty'] as int);
      }
    }
    return out;
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
    stockTracked: r.stockTracked,
    variantStock: Map<String, int>.from(jsonDecode(r.variantStockJson)),
  );
}
