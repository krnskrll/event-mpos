import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/db/database.dart';
import '../../core/network/api_client.dart';
import '../../domain/customer.dart';

final memberRepositoryProvider = Provider(
  (ref) => MemberRepository(ref.watch(dioProvider), ref.watch(databaseProvider)),
);

class MemberRepository {
  MemberRepository(this._dio, this._db);

  final Dio _dio;
  final AppDatabase _db;

  /// A member QR: the local cache first (instant, offline), then the server. Null when unknown.
  Future<Customer?> identify(String qrCode) async {
    final code = qrCode.trim();
    final local = await (_db.select(_db.members)..where((m) => m.qrCode.equals(code))).getSingleOrNull();
    if (local != null) {
      return Customer(name: local.name, tier: local.tier, qrCode: local.qrCode, source: CustomerSource.localCache);
    }
    try {
      final m = (await _dio.get('api/me/members/by-qr/${Uri.encodeComponent(code)}')).data;
      return Customer(name: m['name'], tier: m['tier'], qrCode: m['qrCode'], source: CustomerSource.onlineLookup);
    } on DioException {
      return null;
    }
  }
}
