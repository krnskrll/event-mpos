import 'package:drift/drift.dart';

@DataClassName('EventSnapshotRow')
class EventSnapshots extends Table {
  @override
  String get tableName => 'event_snapshot';

  TextColumn get eventId => text()();
  TextColumn get version => text()();
  TextColumn get snapshotJson => text()();
  DateTimeColumn get syncedAt => dateTime()();

  @override
  Set<Column> get primaryKey => {eventId};
}

@DataClassName('ProductRow')
@TableIndex(name: 'idx_product_event', columns: {#eventId})
class Products extends Table {
  @override
  String get tableName => 'product';

  TextColumn get eventId => text()();
  TextColumn get id => text()();
  TextColumn get sku => text()();
  TextColumn get name => text()();
  TextColumn get brand => text().nullable()();
  RealColumn get price => real()();
  TextColumn get imageUrl => text().nullable()();
  TextColumn get barcodes => text()(); // newline-joined
  IntColumn get remainingStock => integer()();
  TextColumn get variantsJson => text().withDefault(const Constant('[]'))();

  @override
  Set<Column> get primaryKey => {eventId, id};
}

@DataClassName('MemberRow')
class Members extends Table {
  @override
  String get tableName => 'member';

  TextColumn get qrCode => text()();
  TextColumn get name => text()();
  TextColumn get tier => text()();

  @override
  Set<Column> get primaryKey => {qrCode};
}

@DataClassName('PromotionRow')
@TableIndex(name: 'idx_promotion_event', columns: {#eventId})
class Promotions extends Table {
  @override
  String get tableName => 'promotion';

  TextColumn get eventId => text()();
  TextColumn get id => text()();
  TextColumn get name => text()();
  TextColumn get type => text()();
  RealColumn get percentOff => real().nullable()();
  RealColumn get amountOff => real().nullable()();
  TextColumn get scope => text()();
  TextColumn get productIds => text()(); // comma-joined
  IntColumn get minQty => integer().nullable()();
  RealColumn get minSubtotal => real().nullable()();
  TextColumn get startsAt => text().nullable()();
  TextColumn get endsAt => text().nullable()();
  IntColumn get priority => integer().withDefault(const Constant(0))();

  @override
  Set<Column> get primaryKey => {eventId, id};
}

@DataClassName('PendingOrderRow')
class PendingOrders extends Table {
  @override
  String get tableName => 'pending_order';

  TextColumn get idempotencyKey => text()();
  TextColumn get eventId => text()();
  TextColumn get customerName => text()();
  TextColumn get customerQr => text().nullable()();
  TextColumn get customerTier => text().nullable()();
  TextColumn get linesJson => text()();
  TextColumn get status => text()(); // PENDING | CONFIRMED | FAILED
  DateTimeColumn get createdAt => dateTime()();
  TextColumn get orderNumberDisplay => text().nullable()();
  TextColumn get referenceId => text().nullable()();
  RealColumn get total => real().nullable()();
  TextColumn get errorMessage => text().nullable()();
  IntColumn get orderNumber => integer().nullable()();
  TextColumn get serverOrderId => text().nullable()();

  @override
  Set<Column> get primaryKey => {idempotencyKey};
}

@DataClassName('ReceiptRow')
class Receipts extends Table {
  @override
  String get tableName => 'receipt';

  TextColumn get orderKey => text()();
  TextColumn get eventId => text()();
  TextColumn get orderNumberDisplay => text()();
  TextColumn get receiptText => text()(); // newline-joined 58 mm lines
  BoolColumn get printedOk => boolean()();
  DateTimeColumn get createdAt => dateTime()();
  TextColumn get backendOrderId => text().nullable()();

  @override
  Set<Column> get primaryKey => {orderKey};
}

@DataClassName('CartDraftRow')
@TableIndex(name: 'idx_cart_draft_event', columns: {#eventId})
class CartDrafts extends Table {
  @override
  String get tableName => 'cart_draft';

  TextColumn get id => text()();
  TextColumn get eventId => text()();
  TextColumn get label => text()();
  TextColumn get customerJson => text().nullable()();
  TextColumn get linesJson => text()();
  DateTimeColumn get createdAt => dateTime()();
  BoolColumn get live => boolean().withDefault(const Constant(false))(); // autosaved current cart

  @override
  Set<Column> get primaryKey => {id};
}

@DataClassName('NumberBlockRow')
class NumberBlocks extends Table {
  @override
  String get tableName => 'number_block';

  TextColumn get eventId => text()();
  TextColumn get businessDate => text()(); // yyyyMMdd
  IntColumn get rangeStart => integer()();
  IntColumn get rangeEnd => integer()();
  IntColumn get nextUnused => integer()();

  @override
  Set<Column> get primaryKey => {eventId, rangeStart};
}
