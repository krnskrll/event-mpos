// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'database.dart';

// ignore_for_file: type=lint
class $EventSnapshotsTable extends EventSnapshots with TableInfo<$EventSnapshotsTable, EventSnapshotRow> {
  @override
  final GeneratedDatabase attachedDatabase;
  final String? _alias;
  $EventSnapshotsTable(this.attachedDatabase, [this._alias]);
  static const VerificationMeta _eventIdMeta = const VerificationMeta('eventId');
  @override
  late final GeneratedColumn<String> eventId = GeneratedColumn<String>(
    'event_id',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _versionMeta = const VerificationMeta('version');
  @override
  late final GeneratedColumn<String> version = GeneratedColumn<String>(
    'version',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _snapshotJsonMeta = const VerificationMeta('snapshotJson');
  @override
  late final GeneratedColumn<String> snapshotJson = GeneratedColumn<String>(
    'snapshot_json',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _syncedAtMeta = const VerificationMeta('syncedAt');
  @override
  late final GeneratedColumn<DateTime> syncedAt = GeneratedColumn<DateTime>(
    'synced_at',
    aliasedName,
    false,
    type: DriftSqlType.dateTime,
    requiredDuringInsert: true,
  );
  @override
  List<GeneratedColumn> get $columns => [eventId, version, snapshotJson, syncedAt];
  @override
  String get aliasedName => _alias ?? actualTableName;
  @override
  String get actualTableName => $name;
  static const String $name = 'event_snapshot';
  @override
  VerificationContext validateIntegrity(Insertable<EventSnapshotRow> instance, {bool isInserting = false}) {
    final context = VerificationContext();
    final data = instance.toColumns(true);
    if (data.containsKey('event_id')) {
      context.handle(_eventIdMeta, eventId.isAcceptableOrUnknown(data['event_id']!, _eventIdMeta));
    } else if (isInserting) {
      context.missing(_eventIdMeta);
    }
    if (data.containsKey('version')) {
      context.handle(_versionMeta, version.isAcceptableOrUnknown(data['version']!, _versionMeta));
    } else if (isInserting) {
      context.missing(_versionMeta);
    }
    if (data.containsKey('snapshot_json')) {
      context.handle(_snapshotJsonMeta, snapshotJson.isAcceptableOrUnknown(data['snapshot_json']!, _snapshotJsonMeta));
    } else if (isInserting) {
      context.missing(_snapshotJsonMeta);
    }
    if (data.containsKey('synced_at')) {
      context.handle(_syncedAtMeta, syncedAt.isAcceptableOrUnknown(data['synced_at']!, _syncedAtMeta));
    } else if (isInserting) {
      context.missing(_syncedAtMeta);
    }
    return context;
  }

  @override
  Set<GeneratedColumn> get $primaryKey => {eventId};
  @override
  EventSnapshotRow map(Map<String, dynamic> data, {String? tablePrefix}) {
    final effectivePrefix = tablePrefix != null ? '$tablePrefix.' : '';
    return EventSnapshotRow(
      eventId: attachedDatabase.typeMapping.read(DriftSqlType.string, data['${effectivePrefix}event_id'])!,
      version: attachedDatabase.typeMapping.read(DriftSqlType.string, data['${effectivePrefix}version'])!,
      snapshotJson: attachedDatabase.typeMapping.read(DriftSqlType.string, data['${effectivePrefix}snapshot_json'])!,
      syncedAt: attachedDatabase.typeMapping.read(DriftSqlType.dateTime, data['${effectivePrefix}synced_at'])!,
    );
  }

  @override
  $EventSnapshotsTable createAlias(String alias) {
    return $EventSnapshotsTable(attachedDatabase, alias);
  }
}

class EventSnapshotRow extends DataClass implements Insertable<EventSnapshotRow> {
  final String eventId;
  final String version;
  final String snapshotJson;
  final DateTime syncedAt;
  const EventSnapshotRow({
    required this.eventId,
    required this.version,
    required this.snapshotJson,
    required this.syncedAt,
  });
  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    map['event_id'] = Variable<String>(eventId);
    map['version'] = Variable<String>(version);
    map['snapshot_json'] = Variable<String>(snapshotJson);
    map['synced_at'] = Variable<DateTime>(syncedAt);
    return map;
  }

  EventSnapshotsCompanion toCompanion(bool nullToAbsent) {
    return EventSnapshotsCompanion(
      eventId: Value(eventId),
      version: Value(version),
      snapshotJson: Value(snapshotJson),
      syncedAt: Value(syncedAt),
    );
  }

  factory EventSnapshotRow.fromJson(Map<String, dynamic> json, {ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return EventSnapshotRow(
      eventId: serializer.fromJson<String>(json['eventId']),
      version: serializer.fromJson<String>(json['version']),
      snapshotJson: serializer.fromJson<String>(json['snapshotJson']),
      syncedAt: serializer.fromJson<DateTime>(json['syncedAt']),
    );
  }
  @override
  Map<String, dynamic> toJson({ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return <String, dynamic>{
      'eventId': serializer.toJson<String>(eventId),
      'version': serializer.toJson<String>(version),
      'snapshotJson': serializer.toJson<String>(snapshotJson),
      'syncedAt': serializer.toJson<DateTime>(syncedAt),
    };
  }

  EventSnapshotRow copyWith({String? eventId, String? version, String? snapshotJson, DateTime? syncedAt}) =>
      EventSnapshotRow(
        eventId: eventId ?? this.eventId,
        version: version ?? this.version,
        snapshotJson: snapshotJson ?? this.snapshotJson,
        syncedAt: syncedAt ?? this.syncedAt,
      );
  EventSnapshotRow copyWithCompanion(EventSnapshotsCompanion data) {
    return EventSnapshotRow(
      eventId: data.eventId.present ? data.eventId.value : this.eventId,
      version: data.version.present ? data.version.value : this.version,
      snapshotJson: data.snapshotJson.present ? data.snapshotJson.value : this.snapshotJson,
      syncedAt: data.syncedAt.present ? data.syncedAt.value : this.syncedAt,
    );
  }

  @override
  String toString() {
    return (StringBuffer('EventSnapshotRow(')
          ..write('eventId: $eventId, ')
          ..write('version: $version, ')
          ..write('snapshotJson: $snapshotJson, ')
          ..write('syncedAt: $syncedAt')
          ..write(')'))
        .toString();
  }

  @override
  int get hashCode => Object.hash(eventId, version, snapshotJson, syncedAt);
  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      (other is EventSnapshotRow &&
          other.eventId == this.eventId &&
          other.version == this.version &&
          other.snapshotJson == this.snapshotJson &&
          other.syncedAt == this.syncedAt);
}

class EventSnapshotsCompanion extends UpdateCompanion<EventSnapshotRow> {
  final Value<String> eventId;
  final Value<String> version;
  final Value<String> snapshotJson;
  final Value<DateTime> syncedAt;
  final Value<int> rowid;
  const EventSnapshotsCompanion({
    this.eventId = const Value.absent(),
    this.version = const Value.absent(),
    this.snapshotJson = const Value.absent(),
    this.syncedAt = const Value.absent(),
    this.rowid = const Value.absent(),
  });
  EventSnapshotsCompanion.insert({
    required String eventId,
    required String version,
    required String snapshotJson,
    required DateTime syncedAt,
    this.rowid = const Value.absent(),
  }) : eventId = Value(eventId),
       version = Value(version),
       snapshotJson = Value(snapshotJson),
       syncedAt = Value(syncedAt);
  static Insertable<EventSnapshotRow> custom({
    Expression<String>? eventId,
    Expression<String>? version,
    Expression<String>? snapshotJson,
    Expression<DateTime>? syncedAt,
    Expression<int>? rowid,
  }) {
    return RawValuesInsertable({
      if (eventId != null) 'event_id': eventId,
      if (version != null) 'version': version,
      if (snapshotJson != null) 'snapshot_json': snapshotJson,
      if (syncedAt != null) 'synced_at': syncedAt,
      if (rowid != null) 'rowid': rowid,
    });
  }

  EventSnapshotsCompanion copyWith({
    Value<String>? eventId,
    Value<String>? version,
    Value<String>? snapshotJson,
    Value<DateTime>? syncedAt,
    Value<int>? rowid,
  }) {
    return EventSnapshotsCompanion(
      eventId: eventId ?? this.eventId,
      version: version ?? this.version,
      snapshotJson: snapshotJson ?? this.snapshotJson,
      syncedAt: syncedAt ?? this.syncedAt,
      rowid: rowid ?? this.rowid,
    );
  }

  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    if (eventId.present) {
      map['event_id'] = Variable<String>(eventId.value);
    }
    if (version.present) {
      map['version'] = Variable<String>(version.value);
    }
    if (snapshotJson.present) {
      map['snapshot_json'] = Variable<String>(snapshotJson.value);
    }
    if (syncedAt.present) {
      map['synced_at'] = Variable<DateTime>(syncedAt.value);
    }
    if (rowid.present) {
      map['rowid'] = Variable<int>(rowid.value);
    }
    return map;
  }

  @override
  String toString() {
    return (StringBuffer('EventSnapshotsCompanion(')
          ..write('eventId: $eventId, ')
          ..write('version: $version, ')
          ..write('snapshotJson: $snapshotJson, ')
          ..write('syncedAt: $syncedAt, ')
          ..write('rowid: $rowid')
          ..write(')'))
        .toString();
  }
}

class $ProductsTable extends Products with TableInfo<$ProductsTable, ProductRow> {
  @override
  final GeneratedDatabase attachedDatabase;
  final String? _alias;
  $ProductsTable(this.attachedDatabase, [this._alias]);
  static const VerificationMeta _eventIdMeta = const VerificationMeta('eventId');
  @override
  late final GeneratedColumn<String> eventId = GeneratedColumn<String>(
    'event_id',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _idMeta = const VerificationMeta('id');
  @override
  late final GeneratedColumn<String> id = GeneratedColumn<String>(
    'id',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _skuMeta = const VerificationMeta('sku');
  @override
  late final GeneratedColumn<String> sku = GeneratedColumn<String>(
    'sku',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _nameMeta = const VerificationMeta('name');
  @override
  late final GeneratedColumn<String> name = GeneratedColumn<String>(
    'name',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _brandMeta = const VerificationMeta('brand');
  @override
  late final GeneratedColumn<String> brand = GeneratedColumn<String>(
    'brand',
    aliasedName,
    true,
    type: DriftSqlType.string,
    requiredDuringInsert: false,
  );
  static const VerificationMeta _priceMeta = const VerificationMeta('price');
  @override
  late final GeneratedColumn<double> price = GeneratedColumn<double>(
    'price',
    aliasedName,
    false,
    type: DriftSqlType.double,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _imageUrlMeta = const VerificationMeta('imageUrl');
  @override
  late final GeneratedColumn<String> imageUrl = GeneratedColumn<String>(
    'image_url',
    aliasedName,
    true,
    type: DriftSqlType.string,
    requiredDuringInsert: false,
  );
  static const VerificationMeta _barcodesMeta = const VerificationMeta('barcodes');
  @override
  late final GeneratedColumn<String> barcodes = GeneratedColumn<String>(
    'barcodes',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _remainingStockMeta = const VerificationMeta('remainingStock');
  @override
  late final GeneratedColumn<int> remainingStock = GeneratedColumn<int>(
    'remaining_stock',
    aliasedName,
    false,
    type: DriftSqlType.int,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _variantsJsonMeta = const VerificationMeta('variantsJson');
  @override
  late final GeneratedColumn<String> variantsJson = GeneratedColumn<String>(
    'variants_json',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: false,
    defaultValue: const Constant('[]'),
  );
  @override
  List<GeneratedColumn> get $columns => [
    eventId,
    id,
    sku,
    name,
    brand,
    price,
    imageUrl,
    barcodes,
    remainingStock,
    variantsJson,
  ];
  @override
  String get aliasedName => _alias ?? actualTableName;
  @override
  String get actualTableName => $name;
  static const String $name = 'product';
  @override
  VerificationContext validateIntegrity(Insertable<ProductRow> instance, {bool isInserting = false}) {
    final context = VerificationContext();
    final data = instance.toColumns(true);
    if (data.containsKey('event_id')) {
      context.handle(_eventIdMeta, eventId.isAcceptableOrUnknown(data['event_id']!, _eventIdMeta));
    } else if (isInserting) {
      context.missing(_eventIdMeta);
    }
    if (data.containsKey('id')) {
      context.handle(_idMeta, id.isAcceptableOrUnknown(data['id']!, _idMeta));
    } else if (isInserting) {
      context.missing(_idMeta);
    }
    if (data.containsKey('sku')) {
      context.handle(_skuMeta, sku.isAcceptableOrUnknown(data['sku']!, _skuMeta));
    } else if (isInserting) {
      context.missing(_skuMeta);
    }
    if (data.containsKey('name')) {
      context.handle(_nameMeta, name.isAcceptableOrUnknown(data['name']!, _nameMeta));
    } else if (isInserting) {
      context.missing(_nameMeta);
    }
    if (data.containsKey('brand')) {
      context.handle(_brandMeta, brand.isAcceptableOrUnknown(data['brand']!, _brandMeta));
    }
    if (data.containsKey('price')) {
      context.handle(_priceMeta, price.isAcceptableOrUnknown(data['price']!, _priceMeta));
    } else if (isInserting) {
      context.missing(_priceMeta);
    }
    if (data.containsKey('image_url')) {
      context.handle(_imageUrlMeta, imageUrl.isAcceptableOrUnknown(data['image_url']!, _imageUrlMeta));
    }
    if (data.containsKey('barcodes')) {
      context.handle(_barcodesMeta, barcodes.isAcceptableOrUnknown(data['barcodes']!, _barcodesMeta));
    } else if (isInserting) {
      context.missing(_barcodesMeta);
    }
    if (data.containsKey('remaining_stock')) {
      context.handle(
        _remainingStockMeta,
        remainingStock.isAcceptableOrUnknown(data['remaining_stock']!, _remainingStockMeta),
      );
    } else if (isInserting) {
      context.missing(_remainingStockMeta);
    }
    if (data.containsKey('variants_json')) {
      context.handle(_variantsJsonMeta, variantsJson.isAcceptableOrUnknown(data['variants_json']!, _variantsJsonMeta));
    }
    return context;
  }

  @override
  Set<GeneratedColumn> get $primaryKey => {eventId, id};
  @override
  ProductRow map(Map<String, dynamic> data, {String? tablePrefix}) {
    final effectivePrefix = tablePrefix != null ? '$tablePrefix.' : '';
    return ProductRow(
      eventId: attachedDatabase.typeMapping.read(DriftSqlType.string, data['${effectivePrefix}event_id'])!,
      id: attachedDatabase.typeMapping.read(DriftSqlType.string, data['${effectivePrefix}id'])!,
      sku: attachedDatabase.typeMapping.read(DriftSqlType.string, data['${effectivePrefix}sku'])!,
      name: attachedDatabase.typeMapping.read(DriftSqlType.string, data['${effectivePrefix}name'])!,
      brand: attachedDatabase.typeMapping.read(DriftSqlType.string, data['${effectivePrefix}brand']),
      price: attachedDatabase.typeMapping.read(DriftSqlType.double, data['${effectivePrefix}price'])!,
      imageUrl: attachedDatabase.typeMapping.read(DriftSqlType.string, data['${effectivePrefix}image_url']),
      barcodes: attachedDatabase.typeMapping.read(DriftSqlType.string, data['${effectivePrefix}barcodes'])!,
      remainingStock: attachedDatabase.typeMapping.read(DriftSqlType.int, data['${effectivePrefix}remaining_stock'])!,
      variantsJson: attachedDatabase.typeMapping.read(DriftSqlType.string, data['${effectivePrefix}variants_json'])!,
    );
  }

  @override
  $ProductsTable createAlias(String alias) {
    return $ProductsTable(attachedDatabase, alias);
  }
}

class ProductRow extends DataClass implements Insertable<ProductRow> {
  final String eventId;
  final String id;
  final String sku;
  final String name;
  final String? brand;
  final double price;
  final String? imageUrl;
  final String barcodes;
  final int remainingStock;
  final String variantsJson;
  const ProductRow({
    required this.eventId,
    required this.id,
    required this.sku,
    required this.name,
    this.brand,
    required this.price,
    this.imageUrl,
    required this.barcodes,
    required this.remainingStock,
    required this.variantsJson,
  });
  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    map['event_id'] = Variable<String>(eventId);
    map['id'] = Variable<String>(id);
    map['sku'] = Variable<String>(sku);
    map['name'] = Variable<String>(name);
    if (!nullToAbsent || brand != null) {
      map['brand'] = Variable<String>(brand);
    }
    map['price'] = Variable<double>(price);
    if (!nullToAbsent || imageUrl != null) {
      map['image_url'] = Variable<String>(imageUrl);
    }
    map['barcodes'] = Variable<String>(barcodes);
    map['remaining_stock'] = Variable<int>(remainingStock);
    map['variants_json'] = Variable<String>(variantsJson);
    return map;
  }

  ProductsCompanion toCompanion(bool nullToAbsent) {
    return ProductsCompanion(
      eventId: Value(eventId),
      id: Value(id),
      sku: Value(sku),
      name: Value(name),
      brand: brand == null && nullToAbsent ? const Value.absent() : Value(brand),
      price: Value(price),
      imageUrl: imageUrl == null && nullToAbsent ? const Value.absent() : Value(imageUrl),
      barcodes: Value(barcodes),
      remainingStock: Value(remainingStock),
      variantsJson: Value(variantsJson),
    );
  }

  factory ProductRow.fromJson(Map<String, dynamic> json, {ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return ProductRow(
      eventId: serializer.fromJson<String>(json['eventId']),
      id: serializer.fromJson<String>(json['id']),
      sku: serializer.fromJson<String>(json['sku']),
      name: serializer.fromJson<String>(json['name']),
      brand: serializer.fromJson<String?>(json['brand']),
      price: serializer.fromJson<double>(json['price']),
      imageUrl: serializer.fromJson<String?>(json['imageUrl']),
      barcodes: serializer.fromJson<String>(json['barcodes']),
      remainingStock: serializer.fromJson<int>(json['remainingStock']),
      variantsJson: serializer.fromJson<String>(json['variantsJson']),
    );
  }
  @override
  Map<String, dynamic> toJson({ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return <String, dynamic>{
      'eventId': serializer.toJson<String>(eventId),
      'id': serializer.toJson<String>(id),
      'sku': serializer.toJson<String>(sku),
      'name': serializer.toJson<String>(name),
      'brand': serializer.toJson<String?>(brand),
      'price': serializer.toJson<double>(price),
      'imageUrl': serializer.toJson<String?>(imageUrl),
      'barcodes': serializer.toJson<String>(barcodes),
      'remainingStock': serializer.toJson<int>(remainingStock),
      'variantsJson': serializer.toJson<String>(variantsJson),
    };
  }

  ProductRow copyWith({
    String? eventId,
    String? id,
    String? sku,
    String? name,
    Value<String?> brand = const Value.absent(),
    double? price,
    Value<String?> imageUrl = const Value.absent(),
    String? barcodes,
    int? remainingStock,
    String? variantsJson,
  }) => ProductRow(
    eventId: eventId ?? this.eventId,
    id: id ?? this.id,
    sku: sku ?? this.sku,
    name: name ?? this.name,
    brand: brand.present ? brand.value : this.brand,
    price: price ?? this.price,
    imageUrl: imageUrl.present ? imageUrl.value : this.imageUrl,
    barcodes: barcodes ?? this.barcodes,
    remainingStock: remainingStock ?? this.remainingStock,
    variantsJson: variantsJson ?? this.variantsJson,
  );
  ProductRow copyWithCompanion(ProductsCompanion data) {
    return ProductRow(
      eventId: data.eventId.present ? data.eventId.value : this.eventId,
      id: data.id.present ? data.id.value : this.id,
      sku: data.sku.present ? data.sku.value : this.sku,
      name: data.name.present ? data.name.value : this.name,
      brand: data.brand.present ? data.brand.value : this.brand,
      price: data.price.present ? data.price.value : this.price,
      imageUrl: data.imageUrl.present ? data.imageUrl.value : this.imageUrl,
      barcodes: data.barcodes.present ? data.barcodes.value : this.barcodes,
      remainingStock: data.remainingStock.present ? data.remainingStock.value : this.remainingStock,
      variantsJson: data.variantsJson.present ? data.variantsJson.value : this.variantsJson,
    );
  }

  @override
  String toString() {
    return (StringBuffer('ProductRow(')
          ..write('eventId: $eventId, ')
          ..write('id: $id, ')
          ..write('sku: $sku, ')
          ..write('name: $name, ')
          ..write('brand: $brand, ')
          ..write('price: $price, ')
          ..write('imageUrl: $imageUrl, ')
          ..write('barcodes: $barcodes, ')
          ..write('remainingStock: $remainingStock, ')
          ..write('variantsJson: $variantsJson')
          ..write(')'))
        .toString();
  }

  @override
  int get hashCode =>
      Object.hash(eventId, id, sku, name, brand, price, imageUrl, barcodes, remainingStock, variantsJson);
  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      (other is ProductRow &&
          other.eventId == this.eventId &&
          other.id == this.id &&
          other.sku == this.sku &&
          other.name == this.name &&
          other.brand == this.brand &&
          other.price == this.price &&
          other.imageUrl == this.imageUrl &&
          other.barcodes == this.barcodes &&
          other.remainingStock == this.remainingStock &&
          other.variantsJson == this.variantsJson);
}

class ProductsCompanion extends UpdateCompanion<ProductRow> {
  final Value<String> eventId;
  final Value<String> id;
  final Value<String> sku;
  final Value<String> name;
  final Value<String?> brand;
  final Value<double> price;
  final Value<String?> imageUrl;
  final Value<String> barcodes;
  final Value<int> remainingStock;
  final Value<String> variantsJson;
  final Value<int> rowid;
  const ProductsCompanion({
    this.eventId = const Value.absent(),
    this.id = const Value.absent(),
    this.sku = const Value.absent(),
    this.name = const Value.absent(),
    this.brand = const Value.absent(),
    this.price = const Value.absent(),
    this.imageUrl = const Value.absent(),
    this.barcodes = const Value.absent(),
    this.remainingStock = const Value.absent(),
    this.variantsJson = const Value.absent(),
    this.rowid = const Value.absent(),
  });
  ProductsCompanion.insert({
    required String eventId,
    required String id,
    required String sku,
    required String name,
    this.brand = const Value.absent(),
    required double price,
    this.imageUrl = const Value.absent(),
    required String barcodes,
    required int remainingStock,
    this.variantsJson = const Value.absent(),
    this.rowid = const Value.absent(),
  }) : eventId = Value(eventId),
       id = Value(id),
       sku = Value(sku),
       name = Value(name),
       price = Value(price),
       barcodes = Value(barcodes),
       remainingStock = Value(remainingStock);
  static Insertable<ProductRow> custom({
    Expression<String>? eventId,
    Expression<String>? id,
    Expression<String>? sku,
    Expression<String>? name,
    Expression<String>? brand,
    Expression<double>? price,
    Expression<String>? imageUrl,
    Expression<String>? barcodes,
    Expression<int>? remainingStock,
    Expression<String>? variantsJson,
    Expression<int>? rowid,
  }) {
    return RawValuesInsertable({
      if (eventId != null) 'event_id': eventId,
      if (id != null) 'id': id,
      if (sku != null) 'sku': sku,
      if (name != null) 'name': name,
      if (brand != null) 'brand': brand,
      if (price != null) 'price': price,
      if (imageUrl != null) 'image_url': imageUrl,
      if (barcodes != null) 'barcodes': barcodes,
      if (remainingStock != null) 'remaining_stock': remainingStock,
      if (variantsJson != null) 'variants_json': variantsJson,
      if (rowid != null) 'rowid': rowid,
    });
  }

  ProductsCompanion copyWith({
    Value<String>? eventId,
    Value<String>? id,
    Value<String>? sku,
    Value<String>? name,
    Value<String?>? brand,
    Value<double>? price,
    Value<String?>? imageUrl,
    Value<String>? barcodes,
    Value<int>? remainingStock,
    Value<String>? variantsJson,
    Value<int>? rowid,
  }) {
    return ProductsCompanion(
      eventId: eventId ?? this.eventId,
      id: id ?? this.id,
      sku: sku ?? this.sku,
      name: name ?? this.name,
      brand: brand ?? this.brand,
      price: price ?? this.price,
      imageUrl: imageUrl ?? this.imageUrl,
      barcodes: barcodes ?? this.barcodes,
      remainingStock: remainingStock ?? this.remainingStock,
      variantsJson: variantsJson ?? this.variantsJson,
      rowid: rowid ?? this.rowid,
    );
  }

  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    if (eventId.present) {
      map['event_id'] = Variable<String>(eventId.value);
    }
    if (id.present) {
      map['id'] = Variable<String>(id.value);
    }
    if (sku.present) {
      map['sku'] = Variable<String>(sku.value);
    }
    if (name.present) {
      map['name'] = Variable<String>(name.value);
    }
    if (brand.present) {
      map['brand'] = Variable<String>(brand.value);
    }
    if (price.present) {
      map['price'] = Variable<double>(price.value);
    }
    if (imageUrl.present) {
      map['image_url'] = Variable<String>(imageUrl.value);
    }
    if (barcodes.present) {
      map['barcodes'] = Variable<String>(barcodes.value);
    }
    if (remainingStock.present) {
      map['remaining_stock'] = Variable<int>(remainingStock.value);
    }
    if (variantsJson.present) {
      map['variants_json'] = Variable<String>(variantsJson.value);
    }
    if (rowid.present) {
      map['rowid'] = Variable<int>(rowid.value);
    }
    return map;
  }

  @override
  String toString() {
    return (StringBuffer('ProductsCompanion(')
          ..write('eventId: $eventId, ')
          ..write('id: $id, ')
          ..write('sku: $sku, ')
          ..write('name: $name, ')
          ..write('brand: $brand, ')
          ..write('price: $price, ')
          ..write('imageUrl: $imageUrl, ')
          ..write('barcodes: $barcodes, ')
          ..write('remainingStock: $remainingStock, ')
          ..write('variantsJson: $variantsJson, ')
          ..write('rowid: $rowid')
          ..write(')'))
        .toString();
  }
}

class $MembersTable extends Members with TableInfo<$MembersTable, MemberRow> {
  @override
  final GeneratedDatabase attachedDatabase;
  final String? _alias;
  $MembersTable(this.attachedDatabase, [this._alias]);
  static const VerificationMeta _qrCodeMeta = const VerificationMeta('qrCode');
  @override
  late final GeneratedColumn<String> qrCode = GeneratedColumn<String>(
    'qr_code',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _nameMeta = const VerificationMeta('name');
  @override
  late final GeneratedColumn<String> name = GeneratedColumn<String>(
    'name',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _tierMeta = const VerificationMeta('tier');
  @override
  late final GeneratedColumn<String> tier = GeneratedColumn<String>(
    'tier',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  @override
  List<GeneratedColumn> get $columns => [qrCode, name, tier];
  @override
  String get aliasedName => _alias ?? actualTableName;
  @override
  String get actualTableName => $name;
  static const String $name = 'member';
  @override
  VerificationContext validateIntegrity(Insertable<MemberRow> instance, {bool isInserting = false}) {
    final context = VerificationContext();
    final data = instance.toColumns(true);
    if (data.containsKey('qr_code')) {
      context.handle(_qrCodeMeta, qrCode.isAcceptableOrUnknown(data['qr_code']!, _qrCodeMeta));
    } else if (isInserting) {
      context.missing(_qrCodeMeta);
    }
    if (data.containsKey('name')) {
      context.handle(_nameMeta, name.isAcceptableOrUnknown(data['name']!, _nameMeta));
    } else if (isInserting) {
      context.missing(_nameMeta);
    }
    if (data.containsKey('tier')) {
      context.handle(_tierMeta, tier.isAcceptableOrUnknown(data['tier']!, _tierMeta));
    } else if (isInserting) {
      context.missing(_tierMeta);
    }
    return context;
  }

  @override
  Set<GeneratedColumn> get $primaryKey => {qrCode};
  @override
  MemberRow map(Map<String, dynamic> data, {String? tablePrefix}) {
    final effectivePrefix = tablePrefix != null ? '$tablePrefix.' : '';
    return MemberRow(
      qrCode: attachedDatabase.typeMapping.read(DriftSqlType.string, data['${effectivePrefix}qr_code'])!,
      name: attachedDatabase.typeMapping.read(DriftSqlType.string, data['${effectivePrefix}name'])!,
      tier: attachedDatabase.typeMapping.read(DriftSqlType.string, data['${effectivePrefix}tier'])!,
    );
  }

  @override
  $MembersTable createAlias(String alias) {
    return $MembersTable(attachedDatabase, alias);
  }
}

class MemberRow extends DataClass implements Insertable<MemberRow> {
  final String qrCode;
  final String name;
  final String tier;
  const MemberRow({required this.qrCode, required this.name, required this.tier});
  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    map['qr_code'] = Variable<String>(qrCode);
    map['name'] = Variable<String>(name);
    map['tier'] = Variable<String>(tier);
    return map;
  }

  MembersCompanion toCompanion(bool nullToAbsent) {
    return MembersCompanion(qrCode: Value(qrCode), name: Value(name), tier: Value(tier));
  }

  factory MemberRow.fromJson(Map<String, dynamic> json, {ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return MemberRow(
      qrCode: serializer.fromJson<String>(json['qrCode']),
      name: serializer.fromJson<String>(json['name']),
      tier: serializer.fromJson<String>(json['tier']),
    );
  }
  @override
  Map<String, dynamic> toJson({ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return <String, dynamic>{
      'qrCode': serializer.toJson<String>(qrCode),
      'name': serializer.toJson<String>(name),
      'tier': serializer.toJson<String>(tier),
    };
  }

  MemberRow copyWith({String? qrCode, String? name, String? tier}) =>
      MemberRow(qrCode: qrCode ?? this.qrCode, name: name ?? this.name, tier: tier ?? this.tier);
  MemberRow copyWithCompanion(MembersCompanion data) {
    return MemberRow(
      qrCode: data.qrCode.present ? data.qrCode.value : this.qrCode,
      name: data.name.present ? data.name.value : this.name,
      tier: data.tier.present ? data.tier.value : this.tier,
    );
  }

  @override
  String toString() {
    return (StringBuffer('MemberRow(')
          ..write('qrCode: $qrCode, ')
          ..write('name: $name, ')
          ..write('tier: $tier')
          ..write(')'))
        .toString();
  }

  @override
  int get hashCode => Object.hash(qrCode, name, tier);
  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      (other is MemberRow && other.qrCode == this.qrCode && other.name == this.name && other.tier == this.tier);
}

class MembersCompanion extends UpdateCompanion<MemberRow> {
  final Value<String> qrCode;
  final Value<String> name;
  final Value<String> tier;
  final Value<int> rowid;
  const MembersCompanion({
    this.qrCode = const Value.absent(),
    this.name = const Value.absent(),
    this.tier = const Value.absent(),
    this.rowid = const Value.absent(),
  });
  MembersCompanion.insert({
    required String qrCode,
    required String name,
    required String tier,
    this.rowid = const Value.absent(),
  }) : qrCode = Value(qrCode),
       name = Value(name),
       tier = Value(tier);
  static Insertable<MemberRow> custom({
    Expression<String>? qrCode,
    Expression<String>? name,
    Expression<String>? tier,
    Expression<int>? rowid,
  }) {
    return RawValuesInsertable({
      if (qrCode != null) 'qr_code': qrCode,
      if (name != null) 'name': name,
      if (tier != null) 'tier': tier,
      if (rowid != null) 'rowid': rowid,
    });
  }

  MembersCompanion copyWith({Value<String>? qrCode, Value<String>? name, Value<String>? tier, Value<int>? rowid}) {
    return MembersCompanion(
      qrCode: qrCode ?? this.qrCode,
      name: name ?? this.name,
      tier: tier ?? this.tier,
      rowid: rowid ?? this.rowid,
    );
  }

  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    if (qrCode.present) {
      map['qr_code'] = Variable<String>(qrCode.value);
    }
    if (name.present) {
      map['name'] = Variable<String>(name.value);
    }
    if (tier.present) {
      map['tier'] = Variable<String>(tier.value);
    }
    if (rowid.present) {
      map['rowid'] = Variable<int>(rowid.value);
    }
    return map;
  }

  @override
  String toString() {
    return (StringBuffer('MembersCompanion(')
          ..write('qrCode: $qrCode, ')
          ..write('name: $name, ')
          ..write('tier: $tier, ')
          ..write('rowid: $rowid')
          ..write(')'))
        .toString();
  }
}

class $PromotionsTable extends Promotions with TableInfo<$PromotionsTable, PromotionRow> {
  @override
  final GeneratedDatabase attachedDatabase;
  final String? _alias;
  $PromotionsTable(this.attachedDatabase, [this._alias]);
  static const VerificationMeta _eventIdMeta = const VerificationMeta('eventId');
  @override
  late final GeneratedColumn<String> eventId = GeneratedColumn<String>(
    'event_id',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _idMeta = const VerificationMeta('id');
  @override
  late final GeneratedColumn<String> id = GeneratedColumn<String>(
    'id',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _nameMeta = const VerificationMeta('name');
  @override
  late final GeneratedColumn<String> name = GeneratedColumn<String>(
    'name',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _typeMeta = const VerificationMeta('type');
  @override
  late final GeneratedColumn<String> type = GeneratedColumn<String>(
    'type',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _percentOffMeta = const VerificationMeta('percentOff');
  @override
  late final GeneratedColumn<double> percentOff = GeneratedColumn<double>(
    'percent_off',
    aliasedName,
    true,
    type: DriftSqlType.double,
    requiredDuringInsert: false,
  );
  static const VerificationMeta _amountOffMeta = const VerificationMeta('amountOff');
  @override
  late final GeneratedColumn<double> amountOff = GeneratedColumn<double>(
    'amount_off',
    aliasedName,
    true,
    type: DriftSqlType.double,
    requiredDuringInsert: false,
  );
  static const VerificationMeta _scopeMeta = const VerificationMeta('scope');
  @override
  late final GeneratedColumn<String> scope = GeneratedColumn<String>(
    'scope',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _productIdsMeta = const VerificationMeta('productIds');
  @override
  late final GeneratedColumn<String> productIds = GeneratedColumn<String>(
    'product_ids',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _minQtyMeta = const VerificationMeta('minQty');
  @override
  late final GeneratedColumn<int> minQty = GeneratedColumn<int>(
    'min_qty',
    aliasedName,
    true,
    type: DriftSqlType.int,
    requiredDuringInsert: false,
  );
  static const VerificationMeta _minSubtotalMeta = const VerificationMeta('minSubtotal');
  @override
  late final GeneratedColumn<double> minSubtotal = GeneratedColumn<double>(
    'min_subtotal',
    aliasedName,
    true,
    type: DriftSqlType.double,
    requiredDuringInsert: false,
  );
  static const VerificationMeta _startsAtMeta = const VerificationMeta('startsAt');
  @override
  late final GeneratedColumn<String> startsAt = GeneratedColumn<String>(
    'starts_at',
    aliasedName,
    true,
    type: DriftSqlType.string,
    requiredDuringInsert: false,
  );
  static const VerificationMeta _endsAtMeta = const VerificationMeta('endsAt');
  @override
  late final GeneratedColumn<String> endsAt = GeneratedColumn<String>(
    'ends_at',
    aliasedName,
    true,
    type: DriftSqlType.string,
    requiredDuringInsert: false,
  );
  static const VerificationMeta _priorityMeta = const VerificationMeta('priority');
  @override
  late final GeneratedColumn<int> priority = GeneratedColumn<int>(
    'priority',
    aliasedName,
    false,
    type: DriftSqlType.int,
    requiredDuringInsert: false,
    defaultValue: const Constant(0),
  );
  @override
  List<GeneratedColumn> get $columns => [
    eventId,
    id,
    name,
    type,
    percentOff,
    amountOff,
    scope,
    productIds,
    minQty,
    minSubtotal,
    startsAt,
    endsAt,
    priority,
  ];
  @override
  String get aliasedName => _alias ?? actualTableName;
  @override
  String get actualTableName => $name;
  static const String $name = 'promotion';
  @override
  VerificationContext validateIntegrity(Insertable<PromotionRow> instance, {bool isInserting = false}) {
    final context = VerificationContext();
    final data = instance.toColumns(true);
    if (data.containsKey('event_id')) {
      context.handle(_eventIdMeta, eventId.isAcceptableOrUnknown(data['event_id']!, _eventIdMeta));
    } else if (isInserting) {
      context.missing(_eventIdMeta);
    }
    if (data.containsKey('id')) {
      context.handle(_idMeta, id.isAcceptableOrUnknown(data['id']!, _idMeta));
    } else if (isInserting) {
      context.missing(_idMeta);
    }
    if (data.containsKey('name')) {
      context.handle(_nameMeta, name.isAcceptableOrUnknown(data['name']!, _nameMeta));
    } else if (isInserting) {
      context.missing(_nameMeta);
    }
    if (data.containsKey('type')) {
      context.handle(_typeMeta, type.isAcceptableOrUnknown(data['type']!, _typeMeta));
    } else if (isInserting) {
      context.missing(_typeMeta);
    }
    if (data.containsKey('percent_off')) {
      context.handle(_percentOffMeta, percentOff.isAcceptableOrUnknown(data['percent_off']!, _percentOffMeta));
    }
    if (data.containsKey('amount_off')) {
      context.handle(_amountOffMeta, amountOff.isAcceptableOrUnknown(data['amount_off']!, _amountOffMeta));
    }
    if (data.containsKey('scope')) {
      context.handle(_scopeMeta, scope.isAcceptableOrUnknown(data['scope']!, _scopeMeta));
    } else if (isInserting) {
      context.missing(_scopeMeta);
    }
    if (data.containsKey('product_ids')) {
      context.handle(_productIdsMeta, productIds.isAcceptableOrUnknown(data['product_ids']!, _productIdsMeta));
    } else if (isInserting) {
      context.missing(_productIdsMeta);
    }
    if (data.containsKey('min_qty')) {
      context.handle(_minQtyMeta, minQty.isAcceptableOrUnknown(data['min_qty']!, _minQtyMeta));
    }
    if (data.containsKey('min_subtotal')) {
      context.handle(_minSubtotalMeta, minSubtotal.isAcceptableOrUnknown(data['min_subtotal']!, _minSubtotalMeta));
    }
    if (data.containsKey('starts_at')) {
      context.handle(_startsAtMeta, startsAt.isAcceptableOrUnknown(data['starts_at']!, _startsAtMeta));
    }
    if (data.containsKey('ends_at')) {
      context.handle(_endsAtMeta, endsAt.isAcceptableOrUnknown(data['ends_at']!, _endsAtMeta));
    }
    if (data.containsKey('priority')) {
      context.handle(_priorityMeta, priority.isAcceptableOrUnknown(data['priority']!, _priorityMeta));
    }
    return context;
  }

  @override
  Set<GeneratedColumn> get $primaryKey => {eventId, id};
  @override
  PromotionRow map(Map<String, dynamic> data, {String? tablePrefix}) {
    final effectivePrefix = tablePrefix != null ? '$tablePrefix.' : '';
    return PromotionRow(
      eventId: attachedDatabase.typeMapping.read(DriftSqlType.string, data['${effectivePrefix}event_id'])!,
      id: attachedDatabase.typeMapping.read(DriftSqlType.string, data['${effectivePrefix}id'])!,
      name: attachedDatabase.typeMapping.read(DriftSqlType.string, data['${effectivePrefix}name'])!,
      type: attachedDatabase.typeMapping.read(DriftSqlType.string, data['${effectivePrefix}type'])!,
      percentOff: attachedDatabase.typeMapping.read(DriftSqlType.double, data['${effectivePrefix}percent_off']),
      amountOff: attachedDatabase.typeMapping.read(DriftSqlType.double, data['${effectivePrefix}amount_off']),
      scope: attachedDatabase.typeMapping.read(DriftSqlType.string, data['${effectivePrefix}scope'])!,
      productIds: attachedDatabase.typeMapping.read(DriftSqlType.string, data['${effectivePrefix}product_ids'])!,
      minQty: attachedDatabase.typeMapping.read(DriftSqlType.int, data['${effectivePrefix}min_qty']),
      minSubtotal: attachedDatabase.typeMapping.read(DriftSqlType.double, data['${effectivePrefix}min_subtotal']),
      startsAt: attachedDatabase.typeMapping.read(DriftSqlType.string, data['${effectivePrefix}starts_at']),
      endsAt: attachedDatabase.typeMapping.read(DriftSqlType.string, data['${effectivePrefix}ends_at']),
      priority: attachedDatabase.typeMapping.read(DriftSqlType.int, data['${effectivePrefix}priority'])!,
    );
  }

  @override
  $PromotionsTable createAlias(String alias) {
    return $PromotionsTable(attachedDatabase, alias);
  }
}

class PromotionRow extends DataClass implements Insertable<PromotionRow> {
  final String eventId;
  final String id;
  final String name;
  final String type;
  final double? percentOff;
  final double? amountOff;
  final String scope;
  final String productIds;
  final int? minQty;
  final double? minSubtotal;
  final String? startsAt;
  final String? endsAt;
  final int priority;
  const PromotionRow({
    required this.eventId,
    required this.id,
    required this.name,
    required this.type,
    this.percentOff,
    this.amountOff,
    required this.scope,
    required this.productIds,
    this.minQty,
    this.minSubtotal,
    this.startsAt,
    this.endsAt,
    required this.priority,
  });
  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    map['event_id'] = Variable<String>(eventId);
    map['id'] = Variable<String>(id);
    map['name'] = Variable<String>(name);
    map['type'] = Variable<String>(type);
    if (!nullToAbsent || percentOff != null) {
      map['percent_off'] = Variable<double>(percentOff);
    }
    if (!nullToAbsent || amountOff != null) {
      map['amount_off'] = Variable<double>(amountOff);
    }
    map['scope'] = Variable<String>(scope);
    map['product_ids'] = Variable<String>(productIds);
    if (!nullToAbsent || minQty != null) {
      map['min_qty'] = Variable<int>(minQty);
    }
    if (!nullToAbsent || minSubtotal != null) {
      map['min_subtotal'] = Variable<double>(minSubtotal);
    }
    if (!nullToAbsent || startsAt != null) {
      map['starts_at'] = Variable<String>(startsAt);
    }
    if (!nullToAbsent || endsAt != null) {
      map['ends_at'] = Variable<String>(endsAt);
    }
    map['priority'] = Variable<int>(priority);
    return map;
  }

  PromotionsCompanion toCompanion(bool nullToAbsent) {
    return PromotionsCompanion(
      eventId: Value(eventId),
      id: Value(id),
      name: Value(name),
      type: Value(type),
      percentOff: percentOff == null && nullToAbsent ? const Value.absent() : Value(percentOff),
      amountOff: amountOff == null && nullToAbsent ? const Value.absent() : Value(amountOff),
      scope: Value(scope),
      productIds: Value(productIds),
      minQty: minQty == null && nullToAbsent ? const Value.absent() : Value(minQty),
      minSubtotal: minSubtotal == null && nullToAbsent ? const Value.absent() : Value(minSubtotal),
      startsAt: startsAt == null && nullToAbsent ? const Value.absent() : Value(startsAt),
      endsAt: endsAt == null && nullToAbsent ? const Value.absent() : Value(endsAt),
      priority: Value(priority),
    );
  }

  factory PromotionRow.fromJson(Map<String, dynamic> json, {ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return PromotionRow(
      eventId: serializer.fromJson<String>(json['eventId']),
      id: serializer.fromJson<String>(json['id']),
      name: serializer.fromJson<String>(json['name']),
      type: serializer.fromJson<String>(json['type']),
      percentOff: serializer.fromJson<double?>(json['percentOff']),
      amountOff: serializer.fromJson<double?>(json['amountOff']),
      scope: serializer.fromJson<String>(json['scope']),
      productIds: serializer.fromJson<String>(json['productIds']),
      minQty: serializer.fromJson<int?>(json['minQty']),
      minSubtotal: serializer.fromJson<double?>(json['minSubtotal']),
      startsAt: serializer.fromJson<String?>(json['startsAt']),
      endsAt: serializer.fromJson<String?>(json['endsAt']),
      priority: serializer.fromJson<int>(json['priority']),
    );
  }
  @override
  Map<String, dynamic> toJson({ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return <String, dynamic>{
      'eventId': serializer.toJson<String>(eventId),
      'id': serializer.toJson<String>(id),
      'name': serializer.toJson<String>(name),
      'type': serializer.toJson<String>(type),
      'percentOff': serializer.toJson<double?>(percentOff),
      'amountOff': serializer.toJson<double?>(amountOff),
      'scope': serializer.toJson<String>(scope),
      'productIds': serializer.toJson<String>(productIds),
      'minQty': serializer.toJson<int?>(minQty),
      'minSubtotal': serializer.toJson<double?>(minSubtotal),
      'startsAt': serializer.toJson<String?>(startsAt),
      'endsAt': serializer.toJson<String?>(endsAt),
      'priority': serializer.toJson<int>(priority),
    };
  }

  PromotionRow copyWith({
    String? eventId,
    String? id,
    String? name,
    String? type,
    Value<double?> percentOff = const Value.absent(),
    Value<double?> amountOff = const Value.absent(),
    String? scope,
    String? productIds,
    Value<int?> minQty = const Value.absent(),
    Value<double?> minSubtotal = const Value.absent(),
    Value<String?> startsAt = const Value.absent(),
    Value<String?> endsAt = const Value.absent(),
    int? priority,
  }) => PromotionRow(
    eventId: eventId ?? this.eventId,
    id: id ?? this.id,
    name: name ?? this.name,
    type: type ?? this.type,
    percentOff: percentOff.present ? percentOff.value : this.percentOff,
    amountOff: amountOff.present ? amountOff.value : this.amountOff,
    scope: scope ?? this.scope,
    productIds: productIds ?? this.productIds,
    minQty: minQty.present ? minQty.value : this.minQty,
    minSubtotal: minSubtotal.present ? minSubtotal.value : this.minSubtotal,
    startsAt: startsAt.present ? startsAt.value : this.startsAt,
    endsAt: endsAt.present ? endsAt.value : this.endsAt,
    priority: priority ?? this.priority,
  );
  PromotionRow copyWithCompanion(PromotionsCompanion data) {
    return PromotionRow(
      eventId: data.eventId.present ? data.eventId.value : this.eventId,
      id: data.id.present ? data.id.value : this.id,
      name: data.name.present ? data.name.value : this.name,
      type: data.type.present ? data.type.value : this.type,
      percentOff: data.percentOff.present ? data.percentOff.value : this.percentOff,
      amountOff: data.amountOff.present ? data.amountOff.value : this.amountOff,
      scope: data.scope.present ? data.scope.value : this.scope,
      productIds: data.productIds.present ? data.productIds.value : this.productIds,
      minQty: data.minQty.present ? data.minQty.value : this.minQty,
      minSubtotal: data.minSubtotal.present ? data.minSubtotal.value : this.minSubtotal,
      startsAt: data.startsAt.present ? data.startsAt.value : this.startsAt,
      endsAt: data.endsAt.present ? data.endsAt.value : this.endsAt,
      priority: data.priority.present ? data.priority.value : this.priority,
    );
  }

  @override
  String toString() {
    return (StringBuffer('PromotionRow(')
          ..write('eventId: $eventId, ')
          ..write('id: $id, ')
          ..write('name: $name, ')
          ..write('type: $type, ')
          ..write('percentOff: $percentOff, ')
          ..write('amountOff: $amountOff, ')
          ..write('scope: $scope, ')
          ..write('productIds: $productIds, ')
          ..write('minQty: $minQty, ')
          ..write('minSubtotal: $minSubtotal, ')
          ..write('startsAt: $startsAt, ')
          ..write('endsAt: $endsAt, ')
          ..write('priority: $priority')
          ..write(')'))
        .toString();
  }

  @override
  int get hashCode => Object.hash(
    eventId,
    id,
    name,
    type,
    percentOff,
    amountOff,
    scope,
    productIds,
    minQty,
    minSubtotal,
    startsAt,
    endsAt,
    priority,
  );
  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      (other is PromotionRow &&
          other.eventId == this.eventId &&
          other.id == this.id &&
          other.name == this.name &&
          other.type == this.type &&
          other.percentOff == this.percentOff &&
          other.amountOff == this.amountOff &&
          other.scope == this.scope &&
          other.productIds == this.productIds &&
          other.minQty == this.minQty &&
          other.minSubtotal == this.minSubtotal &&
          other.startsAt == this.startsAt &&
          other.endsAt == this.endsAt &&
          other.priority == this.priority);
}

class PromotionsCompanion extends UpdateCompanion<PromotionRow> {
  final Value<String> eventId;
  final Value<String> id;
  final Value<String> name;
  final Value<String> type;
  final Value<double?> percentOff;
  final Value<double?> amountOff;
  final Value<String> scope;
  final Value<String> productIds;
  final Value<int?> minQty;
  final Value<double?> minSubtotal;
  final Value<String?> startsAt;
  final Value<String?> endsAt;
  final Value<int> priority;
  final Value<int> rowid;
  const PromotionsCompanion({
    this.eventId = const Value.absent(),
    this.id = const Value.absent(),
    this.name = const Value.absent(),
    this.type = const Value.absent(),
    this.percentOff = const Value.absent(),
    this.amountOff = const Value.absent(),
    this.scope = const Value.absent(),
    this.productIds = const Value.absent(),
    this.minQty = const Value.absent(),
    this.minSubtotal = const Value.absent(),
    this.startsAt = const Value.absent(),
    this.endsAt = const Value.absent(),
    this.priority = const Value.absent(),
    this.rowid = const Value.absent(),
  });
  PromotionsCompanion.insert({
    required String eventId,
    required String id,
    required String name,
    required String type,
    this.percentOff = const Value.absent(),
    this.amountOff = const Value.absent(),
    required String scope,
    required String productIds,
    this.minQty = const Value.absent(),
    this.minSubtotal = const Value.absent(),
    this.startsAt = const Value.absent(),
    this.endsAt = const Value.absent(),
    this.priority = const Value.absent(),
    this.rowid = const Value.absent(),
  }) : eventId = Value(eventId),
       id = Value(id),
       name = Value(name),
       type = Value(type),
       scope = Value(scope),
       productIds = Value(productIds);
  static Insertable<PromotionRow> custom({
    Expression<String>? eventId,
    Expression<String>? id,
    Expression<String>? name,
    Expression<String>? type,
    Expression<double>? percentOff,
    Expression<double>? amountOff,
    Expression<String>? scope,
    Expression<String>? productIds,
    Expression<int>? minQty,
    Expression<double>? minSubtotal,
    Expression<String>? startsAt,
    Expression<String>? endsAt,
    Expression<int>? priority,
    Expression<int>? rowid,
  }) {
    return RawValuesInsertable({
      if (eventId != null) 'event_id': eventId,
      if (id != null) 'id': id,
      if (name != null) 'name': name,
      if (type != null) 'type': type,
      if (percentOff != null) 'percent_off': percentOff,
      if (amountOff != null) 'amount_off': amountOff,
      if (scope != null) 'scope': scope,
      if (productIds != null) 'product_ids': productIds,
      if (minQty != null) 'min_qty': minQty,
      if (minSubtotal != null) 'min_subtotal': minSubtotal,
      if (startsAt != null) 'starts_at': startsAt,
      if (endsAt != null) 'ends_at': endsAt,
      if (priority != null) 'priority': priority,
      if (rowid != null) 'rowid': rowid,
    });
  }

  PromotionsCompanion copyWith({
    Value<String>? eventId,
    Value<String>? id,
    Value<String>? name,
    Value<String>? type,
    Value<double?>? percentOff,
    Value<double?>? amountOff,
    Value<String>? scope,
    Value<String>? productIds,
    Value<int?>? minQty,
    Value<double?>? minSubtotal,
    Value<String?>? startsAt,
    Value<String?>? endsAt,
    Value<int>? priority,
    Value<int>? rowid,
  }) {
    return PromotionsCompanion(
      eventId: eventId ?? this.eventId,
      id: id ?? this.id,
      name: name ?? this.name,
      type: type ?? this.type,
      percentOff: percentOff ?? this.percentOff,
      amountOff: amountOff ?? this.amountOff,
      scope: scope ?? this.scope,
      productIds: productIds ?? this.productIds,
      minQty: minQty ?? this.minQty,
      minSubtotal: minSubtotal ?? this.minSubtotal,
      startsAt: startsAt ?? this.startsAt,
      endsAt: endsAt ?? this.endsAt,
      priority: priority ?? this.priority,
      rowid: rowid ?? this.rowid,
    );
  }

  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    if (eventId.present) {
      map['event_id'] = Variable<String>(eventId.value);
    }
    if (id.present) {
      map['id'] = Variable<String>(id.value);
    }
    if (name.present) {
      map['name'] = Variable<String>(name.value);
    }
    if (type.present) {
      map['type'] = Variable<String>(type.value);
    }
    if (percentOff.present) {
      map['percent_off'] = Variable<double>(percentOff.value);
    }
    if (amountOff.present) {
      map['amount_off'] = Variable<double>(amountOff.value);
    }
    if (scope.present) {
      map['scope'] = Variable<String>(scope.value);
    }
    if (productIds.present) {
      map['product_ids'] = Variable<String>(productIds.value);
    }
    if (minQty.present) {
      map['min_qty'] = Variable<int>(minQty.value);
    }
    if (minSubtotal.present) {
      map['min_subtotal'] = Variable<double>(minSubtotal.value);
    }
    if (startsAt.present) {
      map['starts_at'] = Variable<String>(startsAt.value);
    }
    if (endsAt.present) {
      map['ends_at'] = Variable<String>(endsAt.value);
    }
    if (priority.present) {
      map['priority'] = Variable<int>(priority.value);
    }
    if (rowid.present) {
      map['rowid'] = Variable<int>(rowid.value);
    }
    return map;
  }

  @override
  String toString() {
    return (StringBuffer('PromotionsCompanion(')
          ..write('eventId: $eventId, ')
          ..write('id: $id, ')
          ..write('name: $name, ')
          ..write('type: $type, ')
          ..write('percentOff: $percentOff, ')
          ..write('amountOff: $amountOff, ')
          ..write('scope: $scope, ')
          ..write('productIds: $productIds, ')
          ..write('minQty: $minQty, ')
          ..write('minSubtotal: $minSubtotal, ')
          ..write('startsAt: $startsAt, ')
          ..write('endsAt: $endsAt, ')
          ..write('priority: $priority, ')
          ..write('rowid: $rowid')
          ..write(')'))
        .toString();
  }
}

class $PendingOrdersTable extends PendingOrders with TableInfo<$PendingOrdersTable, PendingOrderRow> {
  @override
  final GeneratedDatabase attachedDatabase;
  final String? _alias;
  $PendingOrdersTable(this.attachedDatabase, [this._alias]);
  static const VerificationMeta _idempotencyKeyMeta = const VerificationMeta('idempotencyKey');
  @override
  late final GeneratedColumn<String> idempotencyKey = GeneratedColumn<String>(
    'idempotency_key',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _eventIdMeta = const VerificationMeta('eventId');
  @override
  late final GeneratedColumn<String> eventId = GeneratedColumn<String>(
    'event_id',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _customerNameMeta = const VerificationMeta('customerName');
  @override
  late final GeneratedColumn<String> customerName = GeneratedColumn<String>(
    'customer_name',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _customerQrMeta = const VerificationMeta('customerQr');
  @override
  late final GeneratedColumn<String> customerQr = GeneratedColumn<String>(
    'customer_qr',
    aliasedName,
    true,
    type: DriftSqlType.string,
    requiredDuringInsert: false,
  );
  static const VerificationMeta _customerTierMeta = const VerificationMeta('customerTier');
  @override
  late final GeneratedColumn<String> customerTier = GeneratedColumn<String>(
    'customer_tier',
    aliasedName,
    true,
    type: DriftSqlType.string,
    requiredDuringInsert: false,
  );
  static const VerificationMeta _linesJsonMeta = const VerificationMeta('linesJson');
  @override
  late final GeneratedColumn<String> linesJson = GeneratedColumn<String>(
    'lines_json',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _statusMeta = const VerificationMeta('status');
  @override
  late final GeneratedColumn<String> status = GeneratedColumn<String>(
    'status',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _createdAtMeta = const VerificationMeta('createdAt');
  @override
  late final GeneratedColumn<DateTime> createdAt = GeneratedColumn<DateTime>(
    'created_at',
    aliasedName,
    false,
    type: DriftSqlType.dateTime,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _orderNumberDisplayMeta = const VerificationMeta('orderNumberDisplay');
  @override
  late final GeneratedColumn<String> orderNumberDisplay = GeneratedColumn<String>(
    'order_number_display',
    aliasedName,
    true,
    type: DriftSqlType.string,
    requiredDuringInsert: false,
  );
  static const VerificationMeta _referenceIdMeta = const VerificationMeta('referenceId');
  @override
  late final GeneratedColumn<String> referenceId = GeneratedColumn<String>(
    'reference_id',
    aliasedName,
    true,
    type: DriftSqlType.string,
    requiredDuringInsert: false,
  );
  static const VerificationMeta _totalMeta = const VerificationMeta('total');
  @override
  late final GeneratedColumn<double> total = GeneratedColumn<double>(
    'total',
    aliasedName,
    true,
    type: DriftSqlType.double,
    requiredDuringInsert: false,
  );
  static const VerificationMeta _errorMessageMeta = const VerificationMeta('errorMessage');
  @override
  late final GeneratedColumn<String> errorMessage = GeneratedColumn<String>(
    'error_message',
    aliasedName,
    true,
    type: DriftSqlType.string,
    requiredDuringInsert: false,
  );
  static const VerificationMeta _orderNumberMeta = const VerificationMeta('orderNumber');
  @override
  late final GeneratedColumn<int> orderNumber = GeneratedColumn<int>(
    'order_number',
    aliasedName,
    true,
    type: DriftSqlType.int,
    requiredDuringInsert: false,
  );
  static const VerificationMeta _serverOrderIdMeta = const VerificationMeta('serverOrderId');
  @override
  late final GeneratedColumn<String> serverOrderId = GeneratedColumn<String>(
    'server_order_id',
    aliasedName,
    true,
    type: DriftSqlType.string,
    requiredDuringInsert: false,
  );
  @override
  List<GeneratedColumn> get $columns => [
    idempotencyKey,
    eventId,
    customerName,
    customerQr,
    customerTier,
    linesJson,
    status,
    createdAt,
    orderNumberDisplay,
    referenceId,
    total,
    errorMessage,
    orderNumber,
    serverOrderId,
  ];
  @override
  String get aliasedName => _alias ?? actualTableName;
  @override
  String get actualTableName => $name;
  static const String $name = 'pending_order';
  @override
  VerificationContext validateIntegrity(Insertable<PendingOrderRow> instance, {bool isInserting = false}) {
    final context = VerificationContext();
    final data = instance.toColumns(true);
    if (data.containsKey('idempotency_key')) {
      context.handle(
        _idempotencyKeyMeta,
        idempotencyKey.isAcceptableOrUnknown(data['idempotency_key']!, _idempotencyKeyMeta),
      );
    } else if (isInserting) {
      context.missing(_idempotencyKeyMeta);
    }
    if (data.containsKey('event_id')) {
      context.handle(_eventIdMeta, eventId.isAcceptableOrUnknown(data['event_id']!, _eventIdMeta));
    } else if (isInserting) {
      context.missing(_eventIdMeta);
    }
    if (data.containsKey('customer_name')) {
      context.handle(_customerNameMeta, customerName.isAcceptableOrUnknown(data['customer_name']!, _customerNameMeta));
    } else if (isInserting) {
      context.missing(_customerNameMeta);
    }
    if (data.containsKey('customer_qr')) {
      context.handle(_customerQrMeta, customerQr.isAcceptableOrUnknown(data['customer_qr']!, _customerQrMeta));
    }
    if (data.containsKey('customer_tier')) {
      context.handle(_customerTierMeta, customerTier.isAcceptableOrUnknown(data['customer_tier']!, _customerTierMeta));
    }
    if (data.containsKey('lines_json')) {
      context.handle(_linesJsonMeta, linesJson.isAcceptableOrUnknown(data['lines_json']!, _linesJsonMeta));
    } else if (isInserting) {
      context.missing(_linesJsonMeta);
    }
    if (data.containsKey('status')) {
      context.handle(_statusMeta, status.isAcceptableOrUnknown(data['status']!, _statusMeta));
    } else if (isInserting) {
      context.missing(_statusMeta);
    }
    if (data.containsKey('created_at')) {
      context.handle(_createdAtMeta, createdAt.isAcceptableOrUnknown(data['created_at']!, _createdAtMeta));
    } else if (isInserting) {
      context.missing(_createdAtMeta);
    }
    if (data.containsKey('order_number_display')) {
      context.handle(
        _orderNumberDisplayMeta,
        orderNumberDisplay.isAcceptableOrUnknown(data['order_number_display']!, _orderNumberDisplayMeta),
      );
    }
    if (data.containsKey('reference_id')) {
      context.handle(_referenceIdMeta, referenceId.isAcceptableOrUnknown(data['reference_id']!, _referenceIdMeta));
    }
    if (data.containsKey('total')) {
      context.handle(_totalMeta, total.isAcceptableOrUnknown(data['total']!, _totalMeta));
    }
    if (data.containsKey('error_message')) {
      context.handle(_errorMessageMeta, errorMessage.isAcceptableOrUnknown(data['error_message']!, _errorMessageMeta));
    }
    if (data.containsKey('order_number')) {
      context.handle(_orderNumberMeta, orderNumber.isAcceptableOrUnknown(data['order_number']!, _orderNumberMeta));
    }
    if (data.containsKey('server_order_id')) {
      context.handle(
        _serverOrderIdMeta,
        serverOrderId.isAcceptableOrUnknown(data['server_order_id']!, _serverOrderIdMeta),
      );
    }
    return context;
  }

  @override
  Set<GeneratedColumn> get $primaryKey => {idempotencyKey};
  @override
  PendingOrderRow map(Map<String, dynamic> data, {String? tablePrefix}) {
    final effectivePrefix = tablePrefix != null ? '$tablePrefix.' : '';
    return PendingOrderRow(
      idempotencyKey: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}idempotency_key'],
      )!,
      eventId: attachedDatabase.typeMapping.read(DriftSqlType.string, data['${effectivePrefix}event_id'])!,
      customerName: attachedDatabase.typeMapping.read(DriftSqlType.string, data['${effectivePrefix}customer_name'])!,
      customerQr: attachedDatabase.typeMapping.read(DriftSqlType.string, data['${effectivePrefix}customer_qr']),
      customerTier: attachedDatabase.typeMapping.read(DriftSqlType.string, data['${effectivePrefix}customer_tier']),
      linesJson: attachedDatabase.typeMapping.read(DriftSqlType.string, data['${effectivePrefix}lines_json'])!,
      status: attachedDatabase.typeMapping.read(DriftSqlType.string, data['${effectivePrefix}status'])!,
      createdAt: attachedDatabase.typeMapping.read(DriftSqlType.dateTime, data['${effectivePrefix}created_at'])!,
      orderNumberDisplay: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}order_number_display'],
      ),
      referenceId: attachedDatabase.typeMapping.read(DriftSqlType.string, data['${effectivePrefix}reference_id']),
      total: attachedDatabase.typeMapping.read(DriftSqlType.double, data['${effectivePrefix}total']),
      errorMessage: attachedDatabase.typeMapping.read(DriftSqlType.string, data['${effectivePrefix}error_message']),
      orderNumber: attachedDatabase.typeMapping.read(DriftSqlType.int, data['${effectivePrefix}order_number']),
      serverOrderId: attachedDatabase.typeMapping.read(DriftSqlType.string, data['${effectivePrefix}server_order_id']),
    );
  }

  @override
  $PendingOrdersTable createAlias(String alias) {
    return $PendingOrdersTable(attachedDatabase, alias);
  }
}

class PendingOrderRow extends DataClass implements Insertable<PendingOrderRow> {
  final String idempotencyKey;
  final String eventId;
  final String customerName;
  final String? customerQr;
  final String? customerTier;
  final String linesJson;
  final String status;
  final DateTime createdAt;
  final String? orderNumberDisplay;
  final String? referenceId;
  final double? total;
  final String? errorMessage;
  final int? orderNumber;
  final String? serverOrderId;
  const PendingOrderRow({
    required this.idempotencyKey,
    required this.eventId,
    required this.customerName,
    this.customerQr,
    this.customerTier,
    required this.linesJson,
    required this.status,
    required this.createdAt,
    this.orderNumberDisplay,
    this.referenceId,
    this.total,
    this.errorMessage,
    this.orderNumber,
    this.serverOrderId,
  });
  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    map['idempotency_key'] = Variable<String>(idempotencyKey);
    map['event_id'] = Variable<String>(eventId);
    map['customer_name'] = Variable<String>(customerName);
    if (!nullToAbsent || customerQr != null) {
      map['customer_qr'] = Variable<String>(customerQr);
    }
    if (!nullToAbsent || customerTier != null) {
      map['customer_tier'] = Variable<String>(customerTier);
    }
    map['lines_json'] = Variable<String>(linesJson);
    map['status'] = Variable<String>(status);
    map['created_at'] = Variable<DateTime>(createdAt);
    if (!nullToAbsent || orderNumberDisplay != null) {
      map['order_number_display'] = Variable<String>(orderNumberDisplay);
    }
    if (!nullToAbsent || referenceId != null) {
      map['reference_id'] = Variable<String>(referenceId);
    }
    if (!nullToAbsent || total != null) {
      map['total'] = Variable<double>(total);
    }
    if (!nullToAbsent || errorMessage != null) {
      map['error_message'] = Variable<String>(errorMessage);
    }
    if (!nullToAbsent || orderNumber != null) {
      map['order_number'] = Variable<int>(orderNumber);
    }
    if (!nullToAbsent || serverOrderId != null) {
      map['server_order_id'] = Variable<String>(serverOrderId);
    }
    return map;
  }

  PendingOrdersCompanion toCompanion(bool nullToAbsent) {
    return PendingOrdersCompanion(
      idempotencyKey: Value(idempotencyKey),
      eventId: Value(eventId),
      customerName: Value(customerName),
      customerQr: customerQr == null && nullToAbsent ? const Value.absent() : Value(customerQr),
      customerTier: customerTier == null && nullToAbsent ? const Value.absent() : Value(customerTier),
      linesJson: Value(linesJson),
      status: Value(status),
      createdAt: Value(createdAt),
      orderNumberDisplay: orderNumberDisplay == null && nullToAbsent ? const Value.absent() : Value(orderNumberDisplay),
      referenceId: referenceId == null && nullToAbsent ? const Value.absent() : Value(referenceId),
      total: total == null && nullToAbsent ? const Value.absent() : Value(total),
      errorMessage: errorMessage == null && nullToAbsent ? const Value.absent() : Value(errorMessage),
      orderNumber: orderNumber == null && nullToAbsent ? const Value.absent() : Value(orderNumber),
      serverOrderId: serverOrderId == null && nullToAbsent ? const Value.absent() : Value(serverOrderId),
    );
  }

  factory PendingOrderRow.fromJson(Map<String, dynamic> json, {ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return PendingOrderRow(
      idempotencyKey: serializer.fromJson<String>(json['idempotencyKey']),
      eventId: serializer.fromJson<String>(json['eventId']),
      customerName: serializer.fromJson<String>(json['customerName']),
      customerQr: serializer.fromJson<String?>(json['customerQr']),
      customerTier: serializer.fromJson<String?>(json['customerTier']),
      linesJson: serializer.fromJson<String>(json['linesJson']),
      status: serializer.fromJson<String>(json['status']),
      createdAt: serializer.fromJson<DateTime>(json['createdAt']),
      orderNumberDisplay: serializer.fromJson<String?>(json['orderNumberDisplay']),
      referenceId: serializer.fromJson<String?>(json['referenceId']),
      total: serializer.fromJson<double?>(json['total']),
      errorMessage: serializer.fromJson<String?>(json['errorMessage']),
      orderNumber: serializer.fromJson<int?>(json['orderNumber']),
      serverOrderId: serializer.fromJson<String?>(json['serverOrderId']),
    );
  }
  @override
  Map<String, dynamic> toJson({ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return <String, dynamic>{
      'idempotencyKey': serializer.toJson<String>(idempotencyKey),
      'eventId': serializer.toJson<String>(eventId),
      'customerName': serializer.toJson<String>(customerName),
      'customerQr': serializer.toJson<String?>(customerQr),
      'customerTier': serializer.toJson<String?>(customerTier),
      'linesJson': serializer.toJson<String>(linesJson),
      'status': serializer.toJson<String>(status),
      'createdAt': serializer.toJson<DateTime>(createdAt),
      'orderNumberDisplay': serializer.toJson<String?>(orderNumberDisplay),
      'referenceId': serializer.toJson<String?>(referenceId),
      'total': serializer.toJson<double?>(total),
      'errorMessage': serializer.toJson<String?>(errorMessage),
      'orderNumber': serializer.toJson<int?>(orderNumber),
      'serverOrderId': serializer.toJson<String?>(serverOrderId),
    };
  }

  PendingOrderRow copyWith({
    String? idempotencyKey,
    String? eventId,
    String? customerName,
    Value<String?> customerQr = const Value.absent(),
    Value<String?> customerTier = const Value.absent(),
    String? linesJson,
    String? status,
    DateTime? createdAt,
    Value<String?> orderNumberDisplay = const Value.absent(),
    Value<String?> referenceId = const Value.absent(),
    Value<double?> total = const Value.absent(),
    Value<String?> errorMessage = const Value.absent(),
    Value<int?> orderNumber = const Value.absent(),
    Value<String?> serverOrderId = const Value.absent(),
  }) => PendingOrderRow(
    idempotencyKey: idempotencyKey ?? this.idempotencyKey,
    eventId: eventId ?? this.eventId,
    customerName: customerName ?? this.customerName,
    customerQr: customerQr.present ? customerQr.value : this.customerQr,
    customerTier: customerTier.present ? customerTier.value : this.customerTier,
    linesJson: linesJson ?? this.linesJson,
    status: status ?? this.status,
    createdAt: createdAt ?? this.createdAt,
    orderNumberDisplay: orderNumberDisplay.present ? orderNumberDisplay.value : this.orderNumberDisplay,
    referenceId: referenceId.present ? referenceId.value : this.referenceId,
    total: total.present ? total.value : this.total,
    errorMessage: errorMessage.present ? errorMessage.value : this.errorMessage,
    orderNumber: orderNumber.present ? orderNumber.value : this.orderNumber,
    serverOrderId: serverOrderId.present ? serverOrderId.value : this.serverOrderId,
  );
  PendingOrderRow copyWithCompanion(PendingOrdersCompanion data) {
    return PendingOrderRow(
      idempotencyKey: data.idempotencyKey.present ? data.idempotencyKey.value : this.idempotencyKey,
      eventId: data.eventId.present ? data.eventId.value : this.eventId,
      customerName: data.customerName.present ? data.customerName.value : this.customerName,
      customerQr: data.customerQr.present ? data.customerQr.value : this.customerQr,
      customerTier: data.customerTier.present ? data.customerTier.value : this.customerTier,
      linesJson: data.linesJson.present ? data.linesJson.value : this.linesJson,
      status: data.status.present ? data.status.value : this.status,
      createdAt: data.createdAt.present ? data.createdAt.value : this.createdAt,
      orderNumberDisplay: data.orderNumberDisplay.present ? data.orderNumberDisplay.value : this.orderNumberDisplay,
      referenceId: data.referenceId.present ? data.referenceId.value : this.referenceId,
      total: data.total.present ? data.total.value : this.total,
      errorMessage: data.errorMessage.present ? data.errorMessage.value : this.errorMessage,
      orderNumber: data.orderNumber.present ? data.orderNumber.value : this.orderNumber,
      serverOrderId: data.serverOrderId.present ? data.serverOrderId.value : this.serverOrderId,
    );
  }

  @override
  String toString() {
    return (StringBuffer('PendingOrderRow(')
          ..write('idempotencyKey: $idempotencyKey, ')
          ..write('eventId: $eventId, ')
          ..write('customerName: $customerName, ')
          ..write('customerQr: $customerQr, ')
          ..write('customerTier: $customerTier, ')
          ..write('linesJson: $linesJson, ')
          ..write('status: $status, ')
          ..write('createdAt: $createdAt, ')
          ..write('orderNumberDisplay: $orderNumberDisplay, ')
          ..write('referenceId: $referenceId, ')
          ..write('total: $total, ')
          ..write('errorMessage: $errorMessage, ')
          ..write('orderNumber: $orderNumber, ')
          ..write('serverOrderId: $serverOrderId')
          ..write(')'))
        .toString();
  }

  @override
  int get hashCode => Object.hash(
    idempotencyKey,
    eventId,
    customerName,
    customerQr,
    customerTier,
    linesJson,
    status,
    createdAt,
    orderNumberDisplay,
    referenceId,
    total,
    errorMessage,
    orderNumber,
    serverOrderId,
  );
  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      (other is PendingOrderRow &&
          other.idempotencyKey == this.idempotencyKey &&
          other.eventId == this.eventId &&
          other.customerName == this.customerName &&
          other.customerQr == this.customerQr &&
          other.customerTier == this.customerTier &&
          other.linesJson == this.linesJson &&
          other.status == this.status &&
          other.createdAt == this.createdAt &&
          other.orderNumberDisplay == this.orderNumberDisplay &&
          other.referenceId == this.referenceId &&
          other.total == this.total &&
          other.errorMessage == this.errorMessage &&
          other.orderNumber == this.orderNumber &&
          other.serverOrderId == this.serverOrderId);
}

class PendingOrdersCompanion extends UpdateCompanion<PendingOrderRow> {
  final Value<String> idempotencyKey;
  final Value<String> eventId;
  final Value<String> customerName;
  final Value<String?> customerQr;
  final Value<String?> customerTier;
  final Value<String> linesJson;
  final Value<String> status;
  final Value<DateTime> createdAt;
  final Value<String?> orderNumberDisplay;
  final Value<String?> referenceId;
  final Value<double?> total;
  final Value<String?> errorMessage;
  final Value<int?> orderNumber;
  final Value<String?> serverOrderId;
  final Value<int> rowid;
  const PendingOrdersCompanion({
    this.idempotencyKey = const Value.absent(),
    this.eventId = const Value.absent(),
    this.customerName = const Value.absent(),
    this.customerQr = const Value.absent(),
    this.customerTier = const Value.absent(),
    this.linesJson = const Value.absent(),
    this.status = const Value.absent(),
    this.createdAt = const Value.absent(),
    this.orderNumberDisplay = const Value.absent(),
    this.referenceId = const Value.absent(),
    this.total = const Value.absent(),
    this.errorMessage = const Value.absent(),
    this.orderNumber = const Value.absent(),
    this.serverOrderId = const Value.absent(),
    this.rowid = const Value.absent(),
  });
  PendingOrdersCompanion.insert({
    required String idempotencyKey,
    required String eventId,
    required String customerName,
    this.customerQr = const Value.absent(),
    this.customerTier = const Value.absent(),
    required String linesJson,
    required String status,
    required DateTime createdAt,
    this.orderNumberDisplay = const Value.absent(),
    this.referenceId = const Value.absent(),
    this.total = const Value.absent(),
    this.errorMessage = const Value.absent(),
    this.orderNumber = const Value.absent(),
    this.serverOrderId = const Value.absent(),
    this.rowid = const Value.absent(),
  }) : idempotencyKey = Value(idempotencyKey),
       eventId = Value(eventId),
       customerName = Value(customerName),
       linesJson = Value(linesJson),
       status = Value(status),
       createdAt = Value(createdAt);
  static Insertable<PendingOrderRow> custom({
    Expression<String>? idempotencyKey,
    Expression<String>? eventId,
    Expression<String>? customerName,
    Expression<String>? customerQr,
    Expression<String>? customerTier,
    Expression<String>? linesJson,
    Expression<String>? status,
    Expression<DateTime>? createdAt,
    Expression<String>? orderNumberDisplay,
    Expression<String>? referenceId,
    Expression<double>? total,
    Expression<String>? errorMessage,
    Expression<int>? orderNumber,
    Expression<String>? serverOrderId,
    Expression<int>? rowid,
  }) {
    return RawValuesInsertable({
      if (idempotencyKey != null) 'idempotency_key': idempotencyKey,
      if (eventId != null) 'event_id': eventId,
      if (customerName != null) 'customer_name': customerName,
      if (customerQr != null) 'customer_qr': customerQr,
      if (customerTier != null) 'customer_tier': customerTier,
      if (linesJson != null) 'lines_json': linesJson,
      if (status != null) 'status': status,
      if (createdAt != null) 'created_at': createdAt,
      if (orderNumberDisplay != null) 'order_number_display': orderNumberDisplay,
      if (referenceId != null) 'reference_id': referenceId,
      if (total != null) 'total': total,
      if (errorMessage != null) 'error_message': errorMessage,
      if (orderNumber != null) 'order_number': orderNumber,
      if (serverOrderId != null) 'server_order_id': serverOrderId,
      if (rowid != null) 'rowid': rowid,
    });
  }

  PendingOrdersCompanion copyWith({
    Value<String>? idempotencyKey,
    Value<String>? eventId,
    Value<String>? customerName,
    Value<String?>? customerQr,
    Value<String?>? customerTier,
    Value<String>? linesJson,
    Value<String>? status,
    Value<DateTime>? createdAt,
    Value<String?>? orderNumberDisplay,
    Value<String?>? referenceId,
    Value<double?>? total,
    Value<String?>? errorMessage,
    Value<int?>? orderNumber,
    Value<String?>? serverOrderId,
    Value<int>? rowid,
  }) {
    return PendingOrdersCompanion(
      idempotencyKey: idempotencyKey ?? this.idempotencyKey,
      eventId: eventId ?? this.eventId,
      customerName: customerName ?? this.customerName,
      customerQr: customerQr ?? this.customerQr,
      customerTier: customerTier ?? this.customerTier,
      linesJson: linesJson ?? this.linesJson,
      status: status ?? this.status,
      createdAt: createdAt ?? this.createdAt,
      orderNumberDisplay: orderNumberDisplay ?? this.orderNumberDisplay,
      referenceId: referenceId ?? this.referenceId,
      total: total ?? this.total,
      errorMessage: errorMessage ?? this.errorMessage,
      orderNumber: orderNumber ?? this.orderNumber,
      serverOrderId: serverOrderId ?? this.serverOrderId,
      rowid: rowid ?? this.rowid,
    );
  }

  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    if (idempotencyKey.present) {
      map['idempotency_key'] = Variable<String>(idempotencyKey.value);
    }
    if (eventId.present) {
      map['event_id'] = Variable<String>(eventId.value);
    }
    if (customerName.present) {
      map['customer_name'] = Variable<String>(customerName.value);
    }
    if (customerQr.present) {
      map['customer_qr'] = Variable<String>(customerQr.value);
    }
    if (customerTier.present) {
      map['customer_tier'] = Variable<String>(customerTier.value);
    }
    if (linesJson.present) {
      map['lines_json'] = Variable<String>(linesJson.value);
    }
    if (status.present) {
      map['status'] = Variable<String>(status.value);
    }
    if (createdAt.present) {
      map['created_at'] = Variable<DateTime>(createdAt.value);
    }
    if (orderNumberDisplay.present) {
      map['order_number_display'] = Variable<String>(orderNumberDisplay.value);
    }
    if (referenceId.present) {
      map['reference_id'] = Variable<String>(referenceId.value);
    }
    if (total.present) {
      map['total'] = Variable<double>(total.value);
    }
    if (errorMessage.present) {
      map['error_message'] = Variable<String>(errorMessage.value);
    }
    if (orderNumber.present) {
      map['order_number'] = Variable<int>(orderNumber.value);
    }
    if (serverOrderId.present) {
      map['server_order_id'] = Variable<String>(serverOrderId.value);
    }
    if (rowid.present) {
      map['rowid'] = Variable<int>(rowid.value);
    }
    return map;
  }

  @override
  String toString() {
    return (StringBuffer('PendingOrdersCompanion(')
          ..write('idempotencyKey: $idempotencyKey, ')
          ..write('eventId: $eventId, ')
          ..write('customerName: $customerName, ')
          ..write('customerQr: $customerQr, ')
          ..write('customerTier: $customerTier, ')
          ..write('linesJson: $linesJson, ')
          ..write('status: $status, ')
          ..write('createdAt: $createdAt, ')
          ..write('orderNumberDisplay: $orderNumberDisplay, ')
          ..write('referenceId: $referenceId, ')
          ..write('total: $total, ')
          ..write('errorMessage: $errorMessage, ')
          ..write('orderNumber: $orderNumber, ')
          ..write('serverOrderId: $serverOrderId, ')
          ..write('rowid: $rowid')
          ..write(')'))
        .toString();
  }
}

class $ReceiptsTable extends Receipts with TableInfo<$ReceiptsTable, ReceiptRow> {
  @override
  final GeneratedDatabase attachedDatabase;
  final String? _alias;
  $ReceiptsTable(this.attachedDatabase, [this._alias]);
  static const VerificationMeta _orderKeyMeta = const VerificationMeta('orderKey');
  @override
  late final GeneratedColumn<String> orderKey = GeneratedColumn<String>(
    'order_key',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _eventIdMeta = const VerificationMeta('eventId');
  @override
  late final GeneratedColumn<String> eventId = GeneratedColumn<String>(
    'event_id',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _orderNumberDisplayMeta = const VerificationMeta('orderNumberDisplay');
  @override
  late final GeneratedColumn<String> orderNumberDisplay = GeneratedColumn<String>(
    'order_number_display',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _receiptTextMeta = const VerificationMeta('receiptText');
  @override
  late final GeneratedColumn<String> receiptText = GeneratedColumn<String>(
    'receipt_text',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _printedOkMeta = const VerificationMeta('printedOk');
  @override
  late final GeneratedColumn<bool> printedOk = GeneratedColumn<bool>(
    'printed_ok',
    aliasedName,
    false,
    type: DriftSqlType.bool,
    requiredDuringInsert: true,
    defaultConstraints: GeneratedColumn.constraintIsAlways('CHECK ("printed_ok" IN (0, 1))'),
  );
  static const VerificationMeta _createdAtMeta = const VerificationMeta('createdAt');
  @override
  late final GeneratedColumn<DateTime> createdAt = GeneratedColumn<DateTime>(
    'created_at',
    aliasedName,
    false,
    type: DriftSqlType.dateTime,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _backendOrderIdMeta = const VerificationMeta('backendOrderId');
  @override
  late final GeneratedColumn<String> backendOrderId = GeneratedColumn<String>(
    'backend_order_id',
    aliasedName,
    true,
    type: DriftSqlType.string,
    requiredDuringInsert: false,
  );
  @override
  List<GeneratedColumn> get $columns => [
    orderKey,
    eventId,
    orderNumberDisplay,
    receiptText,
    printedOk,
    createdAt,
    backendOrderId,
  ];
  @override
  String get aliasedName => _alias ?? actualTableName;
  @override
  String get actualTableName => $name;
  static const String $name = 'receipt';
  @override
  VerificationContext validateIntegrity(Insertable<ReceiptRow> instance, {bool isInserting = false}) {
    final context = VerificationContext();
    final data = instance.toColumns(true);
    if (data.containsKey('order_key')) {
      context.handle(_orderKeyMeta, orderKey.isAcceptableOrUnknown(data['order_key']!, _orderKeyMeta));
    } else if (isInserting) {
      context.missing(_orderKeyMeta);
    }
    if (data.containsKey('event_id')) {
      context.handle(_eventIdMeta, eventId.isAcceptableOrUnknown(data['event_id']!, _eventIdMeta));
    } else if (isInserting) {
      context.missing(_eventIdMeta);
    }
    if (data.containsKey('order_number_display')) {
      context.handle(
        _orderNumberDisplayMeta,
        orderNumberDisplay.isAcceptableOrUnknown(data['order_number_display']!, _orderNumberDisplayMeta),
      );
    } else if (isInserting) {
      context.missing(_orderNumberDisplayMeta);
    }
    if (data.containsKey('receipt_text')) {
      context.handle(_receiptTextMeta, receiptText.isAcceptableOrUnknown(data['receipt_text']!, _receiptTextMeta));
    } else if (isInserting) {
      context.missing(_receiptTextMeta);
    }
    if (data.containsKey('printed_ok')) {
      context.handle(_printedOkMeta, printedOk.isAcceptableOrUnknown(data['printed_ok']!, _printedOkMeta));
    } else if (isInserting) {
      context.missing(_printedOkMeta);
    }
    if (data.containsKey('created_at')) {
      context.handle(_createdAtMeta, createdAt.isAcceptableOrUnknown(data['created_at']!, _createdAtMeta));
    } else if (isInserting) {
      context.missing(_createdAtMeta);
    }
    if (data.containsKey('backend_order_id')) {
      context.handle(
        _backendOrderIdMeta,
        backendOrderId.isAcceptableOrUnknown(data['backend_order_id']!, _backendOrderIdMeta),
      );
    }
    return context;
  }

  @override
  Set<GeneratedColumn> get $primaryKey => {orderKey};
  @override
  ReceiptRow map(Map<String, dynamic> data, {String? tablePrefix}) {
    final effectivePrefix = tablePrefix != null ? '$tablePrefix.' : '';
    return ReceiptRow(
      orderKey: attachedDatabase.typeMapping.read(DriftSqlType.string, data['${effectivePrefix}order_key'])!,
      eventId: attachedDatabase.typeMapping.read(DriftSqlType.string, data['${effectivePrefix}event_id'])!,
      orderNumberDisplay: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}order_number_display'],
      )!,
      receiptText: attachedDatabase.typeMapping.read(DriftSqlType.string, data['${effectivePrefix}receipt_text'])!,
      printedOk: attachedDatabase.typeMapping.read(DriftSqlType.bool, data['${effectivePrefix}printed_ok'])!,
      createdAt: attachedDatabase.typeMapping.read(DriftSqlType.dateTime, data['${effectivePrefix}created_at'])!,
      backendOrderId: attachedDatabase.typeMapping.read(
        DriftSqlType.string,
        data['${effectivePrefix}backend_order_id'],
      ),
    );
  }

  @override
  $ReceiptsTable createAlias(String alias) {
    return $ReceiptsTable(attachedDatabase, alias);
  }
}

class ReceiptRow extends DataClass implements Insertable<ReceiptRow> {
  final String orderKey;
  final String eventId;
  final String orderNumberDisplay;
  final String receiptText;
  final bool printedOk;
  final DateTime createdAt;
  final String? backendOrderId;
  const ReceiptRow({
    required this.orderKey,
    required this.eventId,
    required this.orderNumberDisplay,
    required this.receiptText,
    required this.printedOk,
    required this.createdAt,
    this.backendOrderId,
  });
  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    map['order_key'] = Variable<String>(orderKey);
    map['event_id'] = Variable<String>(eventId);
    map['order_number_display'] = Variable<String>(orderNumberDisplay);
    map['receipt_text'] = Variable<String>(receiptText);
    map['printed_ok'] = Variable<bool>(printedOk);
    map['created_at'] = Variable<DateTime>(createdAt);
    if (!nullToAbsent || backendOrderId != null) {
      map['backend_order_id'] = Variable<String>(backendOrderId);
    }
    return map;
  }

  ReceiptsCompanion toCompanion(bool nullToAbsent) {
    return ReceiptsCompanion(
      orderKey: Value(orderKey),
      eventId: Value(eventId),
      orderNumberDisplay: Value(orderNumberDisplay),
      receiptText: Value(receiptText),
      printedOk: Value(printedOk),
      createdAt: Value(createdAt),
      backendOrderId: backendOrderId == null && nullToAbsent ? const Value.absent() : Value(backendOrderId),
    );
  }

  factory ReceiptRow.fromJson(Map<String, dynamic> json, {ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return ReceiptRow(
      orderKey: serializer.fromJson<String>(json['orderKey']),
      eventId: serializer.fromJson<String>(json['eventId']),
      orderNumberDisplay: serializer.fromJson<String>(json['orderNumberDisplay']),
      receiptText: serializer.fromJson<String>(json['receiptText']),
      printedOk: serializer.fromJson<bool>(json['printedOk']),
      createdAt: serializer.fromJson<DateTime>(json['createdAt']),
      backendOrderId: serializer.fromJson<String?>(json['backendOrderId']),
    );
  }
  @override
  Map<String, dynamic> toJson({ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return <String, dynamic>{
      'orderKey': serializer.toJson<String>(orderKey),
      'eventId': serializer.toJson<String>(eventId),
      'orderNumberDisplay': serializer.toJson<String>(orderNumberDisplay),
      'receiptText': serializer.toJson<String>(receiptText),
      'printedOk': serializer.toJson<bool>(printedOk),
      'createdAt': serializer.toJson<DateTime>(createdAt),
      'backendOrderId': serializer.toJson<String?>(backendOrderId),
    };
  }

  ReceiptRow copyWith({
    String? orderKey,
    String? eventId,
    String? orderNumberDisplay,
    String? receiptText,
    bool? printedOk,
    DateTime? createdAt,
    Value<String?> backendOrderId = const Value.absent(),
  }) => ReceiptRow(
    orderKey: orderKey ?? this.orderKey,
    eventId: eventId ?? this.eventId,
    orderNumberDisplay: orderNumberDisplay ?? this.orderNumberDisplay,
    receiptText: receiptText ?? this.receiptText,
    printedOk: printedOk ?? this.printedOk,
    createdAt: createdAt ?? this.createdAt,
    backendOrderId: backendOrderId.present ? backendOrderId.value : this.backendOrderId,
  );
  ReceiptRow copyWithCompanion(ReceiptsCompanion data) {
    return ReceiptRow(
      orderKey: data.orderKey.present ? data.orderKey.value : this.orderKey,
      eventId: data.eventId.present ? data.eventId.value : this.eventId,
      orderNumberDisplay: data.orderNumberDisplay.present ? data.orderNumberDisplay.value : this.orderNumberDisplay,
      receiptText: data.receiptText.present ? data.receiptText.value : this.receiptText,
      printedOk: data.printedOk.present ? data.printedOk.value : this.printedOk,
      createdAt: data.createdAt.present ? data.createdAt.value : this.createdAt,
      backendOrderId: data.backendOrderId.present ? data.backendOrderId.value : this.backendOrderId,
    );
  }

  @override
  String toString() {
    return (StringBuffer('ReceiptRow(')
          ..write('orderKey: $orderKey, ')
          ..write('eventId: $eventId, ')
          ..write('orderNumberDisplay: $orderNumberDisplay, ')
          ..write('receiptText: $receiptText, ')
          ..write('printedOk: $printedOk, ')
          ..write('createdAt: $createdAt, ')
          ..write('backendOrderId: $backendOrderId')
          ..write(')'))
        .toString();
  }

  @override
  int get hashCode =>
      Object.hash(orderKey, eventId, orderNumberDisplay, receiptText, printedOk, createdAt, backendOrderId);
  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      (other is ReceiptRow &&
          other.orderKey == this.orderKey &&
          other.eventId == this.eventId &&
          other.orderNumberDisplay == this.orderNumberDisplay &&
          other.receiptText == this.receiptText &&
          other.printedOk == this.printedOk &&
          other.createdAt == this.createdAt &&
          other.backendOrderId == this.backendOrderId);
}

class ReceiptsCompanion extends UpdateCompanion<ReceiptRow> {
  final Value<String> orderKey;
  final Value<String> eventId;
  final Value<String> orderNumberDisplay;
  final Value<String> receiptText;
  final Value<bool> printedOk;
  final Value<DateTime> createdAt;
  final Value<String?> backendOrderId;
  final Value<int> rowid;
  const ReceiptsCompanion({
    this.orderKey = const Value.absent(),
    this.eventId = const Value.absent(),
    this.orderNumberDisplay = const Value.absent(),
    this.receiptText = const Value.absent(),
    this.printedOk = const Value.absent(),
    this.createdAt = const Value.absent(),
    this.backendOrderId = const Value.absent(),
    this.rowid = const Value.absent(),
  });
  ReceiptsCompanion.insert({
    required String orderKey,
    required String eventId,
    required String orderNumberDisplay,
    required String receiptText,
    required bool printedOk,
    required DateTime createdAt,
    this.backendOrderId = const Value.absent(),
    this.rowid = const Value.absent(),
  }) : orderKey = Value(orderKey),
       eventId = Value(eventId),
       orderNumberDisplay = Value(orderNumberDisplay),
       receiptText = Value(receiptText),
       printedOk = Value(printedOk),
       createdAt = Value(createdAt);
  static Insertable<ReceiptRow> custom({
    Expression<String>? orderKey,
    Expression<String>? eventId,
    Expression<String>? orderNumberDisplay,
    Expression<String>? receiptText,
    Expression<bool>? printedOk,
    Expression<DateTime>? createdAt,
    Expression<String>? backendOrderId,
    Expression<int>? rowid,
  }) {
    return RawValuesInsertable({
      if (orderKey != null) 'order_key': orderKey,
      if (eventId != null) 'event_id': eventId,
      if (orderNumberDisplay != null) 'order_number_display': orderNumberDisplay,
      if (receiptText != null) 'receipt_text': receiptText,
      if (printedOk != null) 'printed_ok': printedOk,
      if (createdAt != null) 'created_at': createdAt,
      if (backendOrderId != null) 'backend_order_id': backendOrderId,
      if (rowid != null) 'rowid': rowid,
    });
  }

  ReceiptsCompanion copyWith({
    Value<String>? orderKey,
    Value<String>? eventId,
    Value<String>? orderNumberDisplay,
    Value<String>? receiptText,
    Value<bool>? printedOk,
    Value<DateTime>? createdAt,
    Value<String?>? backendOrderId,
    Value<int>? rowid,
  }) {
    return ReceiptsCompanion(
      orderKey: orderKey ?? this.orderKey,
      eventId: eventId ?? this.eventId,
      orderNumberDisplay: orderNumberDisplay ?? this.orderNumberDisplay,
      receiptText: receiptText ?? this.receiptText,
      printedOk: printedOk ?? this.printedOk,
      createdAt: createdAt ?? this.createdAt,
      backendOrderId: backendOrderId ?? this.backendOrderId,
      rowid: rowid ?? this.rowid,
    );
  }

  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    if (orderKey.present) {
      map['order_key'] = Variable<String>(orderKey.value);
    }
    if (eventId.present) {
      map['event_id'] = Variable<String>(eventId.value);
    }
    if (orderNumberDisplay.present) {
      map['order_number_display'] = Variable<String>(orderNumberDisplay.value);
    }
    if (receiptText.present) {
      map['receipt_text'] = Variable<String>(receiptText.value);
    }
    if (printedOk.present) {
      map['printed_ok'] = Variable<bool>(printedOk.value);
    }
    if (createdAt.present) {
      map['created_at'] = Variable<DateTime>(createdAt.value);
    }
    if (backendOrderId.present) {
      map['backend_order_id'] = Variable<String>(backendOrderId.value);
    }
    if (rowid.present) {
      map['rowid'] = Variable<int>(rowid.value);
    }
    return map;
  }

  @override
  String toString() {
    return (StringBuffer('ReceiptsCompanion(')
          ..write('orderKey: $orderKey, ')
          ..write('eventId: $eventId, ')
          ..write('orderNumberDisplay: $orderNumberDisplay, ')
          ..write('receiptText: $receiptText, ')
          ..write('printedOk: $printedOk, ')
          ..write('createdAt: $createdAt, ')
          ..write('backendOrderId: $backendOrderId, ')
          ..write('rowid: $rowid')
          ..write(')'))
        .toString();
  }
}

class $CartDraftsTable extends CartDrafts with TableInfo<$CartDraftsTable, CartDraftRow> {
  @override
  final GeneratedDatabase attachedDatabase;
  final String? _alias;
  $CartDraftsTable(this.attachedDatabase, [this._alias]);
  static const VerificationMeta _idMeta = const VerificationMeta('id');
  @override
  late final GeneratedColumn<String> id = GeneratedColumn<String>(
    'id',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _eventIdMeta = const VerificationMeta('eventId');
  @override
  late final GeneratedColumn<String> eventId = GeneratedColumn<String>(
    'event_id',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _labelMeta = const VerificationMeta('label');
  @override
  late final GeneratedColumn<String> label = GeneratedColumn<String>(
    'label',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _customerJsonMeta = const VerificationMeta('customerJson');
  @override
  late final GeneratedColumn<String> customerJson = GeneratedColumn<String>(
    'customer_json',
    aliasedName,
    true,
    type: DriftSqlType.string,
    requiredDuringInsert: false,
  );
  static const VerificationMeta _linesJsonMeta = const VerificationMeta('linesJson');
  @override
  late final GeneratedColumn<String> linesJson = GeneratedColumn<String>(
    'lines_json',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _createdAtMeta = const VerificationMeta('createdAt');
  @override
  late final GeneratedColumn<DateTime> createdAt = GeneratedColumn<DateTime>(
    'created_at',
    aliasedName,
    false,
    type: DriftSqlType.dateTime,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _liveMeta = const VerificationMeta('live');
  @override
  late final GeneratedColumn<bool> live = GeneratedColumn<bool>(
    'live',
    aliasedName,
    false,
    type: DriftSqlType.bool,
    requiredDuringInsert: false,
    defaultConstraints: GeneratedColumn.constraintIsAlways('CHECK ("live" IN (0, 1))'),
    defaultValue: const Constant(false),
  );
  @override
  List<GeneratedColumn> get $columns => [id, eventId, label, customerJson, linesJson, createdAt, live];
  @override
  String get aliasedName => _alias ?? actualTableName;
  @override
  String get actualTableName => $name;
  static const String $name = 'cart_draft';
  @override
  VerificationContext validateIntegrity(Insertable<CartDraftRow> instance, {bool isInserting = false}) {
    final context = VerificationContext();
    final data = instance.toColumns(true);
    if (data.containsKey('id')) {
      context.handle(_idMeta, id.isAcceptableOrUnknown(data['id']!, _idMeta));
    } else if (isInserting) {
      context.missing(_idMeta);
    }
    if (data.containsKey('event_id')) {
      context.handle(_eventIdMeta, eventId.isAcceptableOrUnknown(data['event_id']!, _eventIdMeta));
    } else if (isInserting) {
      context.missing(_eventIdMeta);
    }
    if (data.containsKey('label')) {
      context.handle(_labelMeta, label.isAcceptableOrUnknown(data['label']!, _labelMeta));
    } else if (isInserting) {
      context.missing(_labelMeta);
    }
    if (data.containsKey('customer_json')) {
      context.handle(_customerJsonMeta, customerJson.isAcceptableOrUnknown(data['customer_json']!, _customerJsonMeta));
    }
    if (data.containsKey('lines_json')) {
      context.handle(_linesJsonMeta, linesJson.isAcceptableOrUnknown(data['lines_json']!, _linesJsonMeta));
    } else if (isInserting) {
      context.missing(_linesJsonMeta);
    }
    if (data.containsKey('created_at')) {
      context.handle(_createdAtMeta, createdAt.isAcceptableOrUnknown(data['created_at']!, _createdAtMeta));
    } else if (isInserting) {
      context.missing(_createdAtMeta);
    }
    if (data.containsKey('live')) {
      context.handle(_liveMeta, live.isAcceptableOrUnknown(data['live']!, _liveMeta));
    }
    return context;
  }

  @override
  Set<GeneratedColumn> get $primaryKey => {id};
  @override
  CartDraftRow map(Map<String, dynamic> data, {String? tablePrefix}) {
    final effectivePrefix = tablePrefix != null ? '$tablePrefix.' : '';
    return CartDraftRow(
      id: attachedDatabase.typeMapping.read(DriftSqlType.string, data['${effectivePrefix}id'])!,
      eventId: attachedDatabase.typeMapping.read(DriftSqlType.string, data['${effectivePrefix}event_id'])!,
      label: attachedDatabase.typeMapping.read(DriftSqlType.string, data['${effectivePrefix}label'])!,
      customerJson: attachedDatabase.typeMapping.read(DriftSqlType.string, data['${effectivePrefix}customer_json']),
      linesJson: attachedDatabase.typeMapping.read(DriftSqlType.string, data['${effectivePrefix}lines_json'])!,
      createdAt: attachedDatabase.typeMapping.read(DriftSqlType.dateTime, data['${effectivePrefix}created_at'])!,
      live: attachedDatabase.typeMapping.read(DriftSqlType.bool, data['${effectivePrefix}live'])!,
    );
  }

  @override
  $CartDraftsTable createAlias(String alias) {
    return $CartDraftsTable(attachedDatabase, alias);
  }
}

class CartDraftRow extends DataClass implements Insertable<CartDraftRow> {
  final String id;
  final String eventId;
  final String label;
  final String? customerJson;
  final String linesJson;
  final DateTime createdAt;
  final bool live;
  const CartDraftRow({
    required this.id,
    required this.eventId,
    required this.label,
    this.customerJson,
    required this.linesJson,
    required this.createdAt,
    required this.live,
  });
  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    map['id'] = Variable<String>(id);
    map['event_id'] = Variable<String>(eventId);
    map['label'] = Variable<String>(label);
    if (!nullToAbsent || customerJson != null) {
      map['customer_json'] = Variable<String>(customerJson);
    }
    map['lines_json'] = Variable<String>(linesJson);
    map['created_at'] = Variable<DateTime>(createdAt);
    map['live'] = Variable<bool>(live);
    return map;
  }

  CartDraftsCompanion toCompanion(bool nullToAbsent) {
    return CartDraftsCompanion(
      id: Value(id),
      eventId: Value(eventId),
      label: Value(label),
      customerJson: customerJson == null && nullToAbsent ? const Value.absent() : Value(customerJson),
      linesJson: Value(linesJson),
      createdAt: Value(createdAt),
      live: Value(live),
    );
  }

  factory CartDraftRow.fromJson(Map<String, dynamic> json, {ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return CartDraftRow(
      id: serializer.fromJson<String>(json['id']),
      eventId: serializer.fromJson<String>(json['eventId']),
      label: serializer.fromJson<String>(json['label']),
      customerJson: serializer.fromJson<String?>(json['customerJson']),
      linesJson: serializer.fromJson<String>(json['linesJson']),
      createdAt: serializer.fromJson<DateTime>(json['createdAt']),
      live: serializer.fromJson<bool>(json['live']),
    );
  }
  @override
  Map<String, dynamic> toJson({ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return <String, dynamic>{
      'id': serializer.toJson<String>(id),
      'eventId': serializer.toJson<String>(eventId),
      'label': serializer.toJson<String>(label),
      'customerJson': serializer.toJson<String?>(customerJson),
      'linesJson': serializer.toJson<String>(linesJson),
      'createdAt': serializer.toJson<DateTime>(createdAt),
      'live': serializer.toJson<bool>(live),
    };
  }

  CartDraftRow copyWith({
    String? id,
    String? eventId,
    String? label,
    Value<String?> customerJson = const Value.absent(),
    String? linesJson,
    DateTime? createdAt,
    bool? live,
  }) => CartDraftRow(
    id: id ?? this.id,
    eventId: eventId ?? this.eventId,
    label: label ?? this.label,
    customerJson: customerJson.present ? customerJson.value : this.customerJson,
    linesJson: linesJson ?? this.linesJson,
    createdAt: createdAt ?? this.createdAt,
    live: live ?? this.live,
  );
  CartDraftRow copyWithCompanion(CartDraftsCompanion data) {
    return CartDraftRow(
      id: data.id.present ? data.id.value : this.id,
      eventId: data.eventId.present ? data.eventId.value : this.eventId,
      label: data.label.present ? data.label.value : this.label,
      customerJson: data.customerJson.present ? data.customerJson.value : this.customerJson,
      linesJson: data.linesJson.present ? data.linesJson.value : this.linesJson,
      createdAt: data.createdAt.present ? data.createdAt.value : this.createdAt,
      live: data.live.present ? data.live.value : this.live,
    );
  }

  @override
  String toString() {
    return (StringBuffer('CartDraftRow(')
          ..write('id: $id, ')
          ..write('eventId: $eventId, ')
          ..write('label: $label, ')
          ..write('customerJson: $customerJson, ')
          ..write('linesJson: $linesJson, ')
          ..write('createdAt: $createdAt, ')
          ..write('live: $live')
          ..write(')'))
        .toString();
  }

  @override
  int get hashCode => Object.hash(id, eventId, label, customerJson, linesJson, createdAt, live);
  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      (other is CartDraftRow &&
          other.id == this.id &&
          other.eventId == this.eventId &&
          other.label == this.label &&
          other.customerJson == this.customerJson &&
          other.linesJson == this.linesJson &&
          other.createdAt == this.createdAt &&
          other.live == this.live);
}

class CartDraftsCompanion extends UpdateCompanion<CartDraftRow> {
  final Value<String> id;
  final Value<String> eventId;
  final Value<String> label;
  final Value<String?> customerJson;
  final Value<String> linesJson;
  final Value<DateTime> createdAt;
  final Value<bool> live;
  final Value<int> rowid;
  const CartDraftsCompanion({
    this.id = const Value.absent(),
    this.eventId = const Value.absent(),
    this.label = const Value.absent(),
    this.customerJson = const Value.absent(),
    this.linesJson = const Value.absent(),
    this.createdAt = const Value.absent(),
    this.live = const Value.absent(),
    this.rowid = const Value.absent(),
  });
  CartDraftsCompanion.insert({
    required String id,
    required String eventId,
    required String label,
    this.customerJson = const Value.absent(),
    required String linesJson,
    required DateTime createdAt,
    this.live = const Value.absent(),
    this.rowid = const Value.absent(),
  }) : id = Value(id),
       eventId = Value(eventId),
       label = Value(label),
       linesJson = Value(linesJson),
       createdAt = Value(createdAt);
  static Insertable<CartDraftRow> custom({
    Expression<String>? id,
    Expression<String>? eventId,
    Expression<String>? label,
    Expression<String>? customerJson,
    Expression<String>? linesJson,
    Expression<DateTime>? createdAt,
    Expression<bool>? live,
    Expression<int>? rowid,
  }) {
    return RawValuesInsertable({
      if (id != null) 'id': id,
      if (eventId != null) 'event_id': eventId,
      if (label != null) 'label': label,
      if (customerJson != null) 'customer_json': customerJson,
      if (linesJson != null) 'lines_json': linesJson,
      if (createdAt != null) 'created_at': createdAt,
      if (live != null) 'live': live,
      if (rowid != null) 'rowid': rowid,
    });
  }

  CartDraftsCompanion copyWith({
    Value<String>? id,
    Value<String>? eventId,
    Value<String>? label,
    Value<String?>? customerJson,
    Value<String>? linesJson,
    Value<DateTime>? createdAt,
    Value<bool>? live,
    Value<int>? rowid,
  }) {
    return CartDraftsCompanion(
      id: id ?? this.id,
      eventId: eventId ?? this.eventId,
      label: label ?? this.label,
      customerJson: customerJson ?? this.customerJson,
      linesJson: linesJson ?? this.linesJson,
      createdAt: createdAt ?? this.createdAt,
      live: live ?? this.live,
      rowid: rowid ?? this.rowid,
    );
  }

  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    if (id.present) {
      map['id'] = Variable<String>(id.value);
    }
    if (eventId.present) {
      map['event_id'] = Variable<String>(eventId.value);
    }
    if (label.present) {
      map['label'] = Variable<String>(label.value);
    }
    if (customerJson.present) {
      map['customer_json'] = Variable<String>(customerJson.value);
    }
    if (linesJson.present) {
      map['lines_json'] = Variable<String>(linesJson.value);
    }
    if (createdAt.present) {
      map['created_at'] = Variable<DateTime>(createdAt.value);
    }
    if (live.present) {
      map['live'] = Variable<bool>(live.value);
    }
    if (rowid.present) {
      map['rowid'] = Variable<int>(rowid.value);
    }
    return map;
  }

  @override
  String toString() {
    return (StringBuffer('CartDraftsCompanion(')
          ..write('id: $id, ')
          ..write('eventId: $eventId, ')
          ..write('label: $label, ')
          ..write('customerJson: $customerJson, ')
          ..write('linesJson: $linesJson, ')
          ..write('createdAt: $createdAt, ')
          ..write('live: $live, ')
          ..write('rowid: $rowid')
          ..write(')'))
        .toString();
  }
}

class $NumberBlocksTable extends NumberBlocks with TableInfo<$NumberBlocksTable, NumberBlockRow> {
  @override
  final GeneratedDatabase attachedDatabase;
  final String? _alias;
  $NumberBlocksTable(this.attachedDatabase, [this._alias]);
  static const VerificationMeta _eventIdMeta = const VerificationMeta('eventId');
  @override
  late final GeneratedColumn<String> eventId = GeneratedColumn<String>(
    'event_id',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _businessDateMeta = const VerificationMeta('businessDate');
  @override
  late final GeneratedColumn<String> businessDate = GeneratedColumn<String>(
    'business_date',
    aliasedName,
    false,
    type: DriftSqlType.string,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _rangeStartMeta = const VerificationMeta('rangeStart');
  @override
  late final GeneratedColumn<int> rangeStart = GeneratedColumn<int>(
    'range_start',
    aliasedName,
    false,
    type: DriftSqlType.int,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _rangeEndMeta = const VerificationMeta('rangeEnd');
  @override
  late final GeneratedColumn<int> rangeEnd = GeneratedColumn<int>(
    'range_end',
    aliasedName,
    false,
    type: DriftSqlType.int,
    requiredDuringInsert: true,
  );
  static const VerificationMeta _nextUnusedMeta = const VerificationMeta('nextUnused');
  @override
  late final GeneratedColumn<int> nextUnused = GeneratedColumn<int>(
    'next_unused',
    aliasedName,
    false,
    type: DriftSqlType.int,
    requiredDuringInsert: true,
  );
  @override
  List<GeneratedColumn> get $columns => [eventId, businessDate, rangeStart, rangeEnd, nextUnused];
  @override
  String get aliasedName => _alias ?? actualTableName;
  @override
  String get actualTableName => $name;
  static const String $name = 'number_block';
  @override
  VerificationContext validateIntegrity(Insertable<NumberBlockRow> instance, {bool isInserting = false}) {
    final context = VerificationContext();
    final data = instance.toColumns(true);
    if (data.containsKey('event_id')) {
      context.handle(_eventIdMeta, eventId.isAcceptableOrUnknown(data['event_id']!, _eventIdMeta));
    } else if (isInserting) {
      context.missing(_eventIdMeta);
    }
    if (data.containsKey('business_date')) {
      context.handle(_businessDateMeta, businessDate.isAcceptableOrUnknown(data['business_date']!, _businessDateMeta));
    } else if (isInserting) {
      context.missing(_businessDateMeta);
    }
    if (data.containsKey('range_start')) {
      context.handle(_rangeStartMeta, rangeStart.isAcceptableOrUnknown(data['range_start']!, _rangeStartMeta));
    } else if (isInserting) {
      context.missing(_rangeStartMeta);
    }
    if (data.containsKey('range_end')) {
      context.handle(_rangeEndMeta, rangeEnd.isAcceptableOrUnknown(data['range_end']!, _rangeEndMeta));
    } else if (isInserting) {
      context.missing(_rangeEndMeta);
    }
    if (data.containsKey('next_unused')) {
      context.handle(_nextUnusedMeta, nextUnused.isAcceptableOrUnknown(data['next_unused']!, _nextUnusedMeta));
    } else if (isInserting) {
      context.missing(_nextUnusedMeta);
    }
    return context;
  }

  @override
  Set<GeneratedColumn> get $primaryKey => {eventId, rangeStart};
  @override
  NumberBlockRow map(Map<String, dynamic> data, {String? tablePrefix}) {
    final effectivePrefix = tablePrefix != null ? '$tablePrefix.' : '';
    return NumberBlockRow(
      eventId: attachedDatabase.typeMapping.read(DriftSqlType.string, data['${effectivePrefix}event_id'])!,
      businessDate: attachedDatabase.typeMapping.read(DriftSqlType.string, data['${effectivePrefix}business_date'])!,
      rangeStart: attachedDatabase.typeMapping.read(DriftSqlType.int, data['${effectivePrefix}range_start'])!,
      rangeEnd: attachedDatabase.typeMapping.read(DriftSqlType.int, data['${effectivePrefix}range_end'])!,
      nextUnused: attachedDatabase.typeMapping.read(DriftSqlType.int, data['${effectivePrefix}next_unused'])!,
    );
  }

  @override
  $NumberBlocksTable createAlias(String alias) {
    return $NumberBlocksTable(attachedDatabase, alias);
  }
}

class NumberBlockRow extends DataClass implements Insertable<NumberBlockRow> {
  final String eventId;
  final String businessDate;
  final int rangeStart;
  final int rangeEnd;
  final int nextUnused;
  const NumberBlockRow({
    required this.eventId,
    required this.businessDate,
    required this.rangeStart,
    required this.rangeEnd,
    required this.nextUnused,
  });
  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    map['event_id'] = Variable<String>(eventId);
    map['business_date'] = Variable<String>(businessDate);
    map['range_start'] = Variable<int>(rangeStart);
    map['range_end'] = Variable<int>(rangeEnd);
    map['next_unused'] = Variable<int>(nextUnused);
    return map;
  }

  NumberBlocksCompanion toCompanion(bool nullToAbsent) {
    return NumberBlocksCompanion(
      eventId: Value(eventId),
      businessDate: Value(businessDate),
      rangeStart: Value(rangeStart),
      rangeEnd: Value(rangeEnd),
      nextUnused: Value(nextUnused),
    );
  }

  factory NumberBlockRow.fromJson(Map<String, dynamic> json, {ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return NumberBlockRow(
      eventId: serializer.fromJson<String>(json['eventId']),
      businessDate: serializer.fromJson<String>(json['businessDate']),
      rangeStart: serializer.fromJson<int>(json['rangeStart']),
      rangeEnd: serializer.fromJson<int>(json['rangeEnd']),
      nextUnused: serializer.fromJson<int>(json['nextUnused']),
    );
  }
  @override
  Map<String, dynamic> toJson({ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return <String, dynamic>{
      'eventId': serializer.toJson<String>(eventId),
      'businessDate': serializer.toJson<String>(businessDate),
      'rangeStart': serializer.toJson<int>(rangeStart),
      'rangeEnd': serializer.toJson<int>(rangeEnd),
      'nextUnused': serializer.toJson<int>(nextUnused),
    };
  }

  NumberBlockRow copyWith({String? eventId, String? businessDate, int? rangeStart, int? rangeEnd, int? nextUnused}) =>
      NumberBlockRow(
        eventId: eventId ?? this.eventId,
        businessDate: businessDate ?? this.businessDate,
        rangeStart: rangeStart ?? this.rangeStart,
        rangeEnd: rangeEnd ?? this.rangeEnd,
        nextUnused: nextUnused ?? this.nextUnused,
      );
  NumberBlockRow copyWithCompanion(NumberBlocksCompanion data) {
    return NumberBlockRow(
      eventId: data.eventId.present ? data.eventId.value : this.eventId,
      businessDate: data.businessDate.present ? data.businessDate.value : this.businessDate,
      rangeStart: data.rangeStart.present ? data.rangeStart.value : this.rangeStart,
      rangeEnd: data.rangeEnd.present ? data.rangeEnd.value : this.rangeEnd,
      nextUnused: data.nextUnused.present ? data.nextUnused.value : this.nextUnused,
    );
  }

  @override
  String toString() {
    return (StringBuffer('NumberBlockRow(')
          ..write('eventId: $eventId, ')
          ..write('businessDate: $businessDate, ')
          ..write('rangeStart: $rangeStart, ')
          ..write('rangeEnd: $rangeEnd, ')
          ..write('nextUnused: $nextUnused')
          ..write(')'))
        .toString();
  }

  @override
  int get hashCode => Object.hash(eventId, businessDate, rangeStart, rangeEnd, nextUnused);
  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      (other is NumberBlockRow &&
          other.eventId == this.eventId &&
          other.businessDate == this.businessDate &&
          other.rangeStart == this.rangeStart &&
          other.rangeEnd == this.rangeEnd &&
          other.nextUnused == this.nextUnused);
}

class NumberBlocksCompanion extends UpdateCompanion<NumberBlockRow> {
  final Value<String> eventId;
  final Value<String> businessDate;
  final Value<int> rangeStart;
  final Value<int> rangeEnd;
  final Value<int> nextUnused;
  final Value<int> rowid;
  const NumberBlocksCompanion({
    this.eventId = const Value.absent(),
    this.businessDate = const Value.absent(),
    this.rangeStart = const Value.absent(),
    this.rangeEnd = const Value.absent(),
    this.nextUnused = const Value.absent(),
    this.rowid = const Value.absent(),
  });
  NumberBlocksCompanion.insert({
    required String eventId,
    required String businessDate,
    required int rangeStart,
    required int rangeEnd,
    required int nextUnused,
    this.rowid = const Value.absent(),
  }) : eventId = Value(eventId),
       businessDate = Value(businessDate),
       rangeStart = Value(rangeStart),
       rangeEnd = Value(rangeEnd),
       nextUnused = Value(nextUnused);
  static Insertable<NumberBlockRow> custom({
    Expression<String>? eventId,
    Expression<String>? businessDate,
    Expression<int>? rangeStart,
    Expression<int>? rangeEnd,
    Expression<int>? nextUnused,
    Expression<int>? rowid,
  }) {
    return RawValuesInsertable({
      if (eventId != null) 'event_id': eventId,
      if (businessDate != null) 'business_date': businessDate,
      if (rangeStart != null) 'range_start': rangeStart,
      if (rangeEnd != null) 'range_end': rangeEnd,
      if (nextUnused != null) 'next_unused': nextUnused,
      if (rowid != null) 'rowid': rowid,
    });
  }

  NumberBlocksCompanion copyWith({
    Value<String>? eventId,
    Value<String>? businessDate,
    Value<int>? rangeStart,
    Value<int>? rangeEnd,
    Value<int>? nextUnused,
    Value<int>? rowid,
  }) {
    return NumberBlocksCompanion(
      eventId: eventId ?? this.eventId,
      businessDate: businessDate ?? this.businessDate,
      rangeStart: rangeStart ?? this.rangeStart,
      rangeEnd: rangeEnd ?? this.rangeEnd,
      nextUnused: nextUnused ?? this.nextUnused,
      rowid: rowid ?? this.rowid,
    );
  }

  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    if (eventId.present) {
      map['event_id'] = Variable<String>(eventId.value);
    }
    if (businessDate.present) {
      map['business_date'] = Variable<String>(businessDate.value);
    }
    if (rangeStart.present) {
      map['range_start'] = Variable<int>(rangeStart.value);
    }
    if (rangeEnd.present) {
      map['range_end'] = Variable<int>(rangeEnd.value);
    }
    if (nextUnused.present) {
      map['next_unused'] = Variable<int>(nextUnused.value);
    }
    if (rowid.present) {
      map['rowid'] = Variable<int>(rowid.value);
    }
    return map;
  }

  @override
  String toString() {
    return (StringBuffer('NumberBlocksCompanion(')
          ..write('eventId: $eventId, ')
          ..write('businessDate: $businessDate, ')
          ..write('rangeStart: $rangeStart, ')
          ..write('rangeEnd: $rangeEnd, ')
          ..write('nextUnused: $nextUnused, ')
          ..write('rowid: $rowid')
          ..write(')'))
        .toString();
  }
}

abstract class _$AppDatabase extends GeneratedDatabase {
  _$AppDatabase(QueryExecutor e) : super(e);
  $AppDatabaseManager get managers => $AppDatabaseManager(this);
  late final $EventSnapshotsTable eventSnapshots = $EventSnapshotsTable(this);
  late final $ProductsTable products = $ProductsTable(this);
  late final $MembersTable members = $MembersTable(this);
  late final $PromotionsTable promotions = $PromotionsTable(this);
  late final $PendingOrdersTable pendingOrders = $PendingOrdersTable(this);
  late final $ReceiptsTable receipts = $ReceiptsTable(this);
  late final $CartDraftsTable cartDrafts = $CartDraftsTable(this);
  late final $NumberBlocksTable numberBlocks = $NumberBlocksTable(this);
  late final Index idxProductEvent = Index('idx_product_event', 'CREATE INDEX idx_product_event ON product (event_id)');
  late final Index idxPromotionEvent = Index(
    'idx_promotion_event',
    'CREATE INDEX idx_promotion_event ON promotion (event_id)',
  );
  late final Index idxCartDraftEvent = Index(
    'idx_cart_draft_event',
    'CREATE INDEX idx_cart_draft_event ON cart_draft (event_id)',
  );
  @override
  Iterable<TableInfo<Table, Object?>> get allTables => allSchemaEntities.whereType<TableInfo<Table, Object?>>();
  @override
  List<DatabaseSchemaEntity> get allSchemaEntities => [
    eventSnapshots,
    products,
    members,
    promotions,
    pendingOrders,
    receipts,
    cartDrafts,
    numberBlocks,
    idxProductEvent,
    idxPromotionEvent,
    idxCartDraftEvent,
  ];
  @override
  DriftDatabaseOptions get options => const DriftDatabaseOptions(storeDateTimeAsText: true);
}

typedef $$EventSnapshotsTableCreateCompanionBuilder = EventSnapshotsCompanion Function({
  required String eventId,
  required String version,
  required String snapshotJson,
  required DateTime syncedAt,
  Value<int> rowid,
});
typedef $$EventSnapshotsTableUpdateCompanionBuilder = EventSnapshotsCompanion Function({
  Value<String> eventId,
  Value<String> version,
  Value<String> snapshotJson,
  Value<DateTime> syncedAt,
  Value<int> rowid,
});

class $$EventSnapshotsTableFilterComposer extends Composer<_$AppDatabase, $EventSnapshotsTable> {
  $$EventSnapshotsTableFilterComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnFilters<String> get eventId =>
      $composableBuilder(column: $table.eventId, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get version =>
      $composableBuilder(column: $table.version, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get snapshotJson =>
      $composableBuilder(column: $table.snapshotJson, builder: (column) => ColumnFilters(column));

  ColumnFilters<DateTime> get syncedAt =>
      $composableBuilder(column: $table.syncedAt, builder: (column) => ColumnFilters(column));
}

class $$EventSnapshotsTableOrderingComposer extends Composer<_$AppDatabase, $EventSnapshotsTable> {
  $$EventSnapshotsTableOrderingComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnOrderings<String> get eventId =>
      $composableBuilder(column: $table.eventId, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get version =>
      $composableBuilder(column: $table.version, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get snapshotJson =>
      $composableBuilder(column: $table.snapshotJson, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<DateTime> get syncedAt =>
      $composableBuilder(column: $table.syncedAt, builder: (column) => ColumnOrderings(column));
}

class $$EventSnapshotsTableAnnotationComposer extends Composer<_$AppDatabase, $EventSnapshotsTable> {
  $$EventSnapshotsTableAnnotationComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  GeneratedColumn<String> get eventId => $composableBuilder(column: $table.eventId, builder: (column) => column);

  GeneratedColumn<String> get version => $composableBuilder(column: $table.version, builder: (column) => column);

  GeneratedColumn<String> get snapshotJson =>
      $composableBuilder(column: $table.snapshotJson, builder: (column) => column);

  GeneratedColumn<DateTime> get syncedAt => $composableBuilder(column: $table.syncedAt, builder: (column) => column);
}

class $$EventSnapshotsTableTableManager
    extends
        RootTableManager<
          _$AppDatabase,
          $EventSnapshotsTable,
          EventSnapshotRow,
          $$EventSnapshotsTableFilterComposer,
          $$EventSnapshotsTableOrderingComposer,
          $$EventSnapshotsTableAnnotationComposer,
          $$EventSnapshotsTableCreateCompanionBuilder,
          $$EventSnapshotsTableUpdateCompanionBuilder,
          (EventSnapshotRow, BaseReferences<_$AppDatabase, $EventSnapshotsTable, EventSnapshotRow>),
          EventSnapshotRow,
          PrefetchHooks Function()
        > {
  $$EventSnapshotsTableTableManager(_$AppDatabase db, $EventSnapshotsTable table)
    : super(
        TableManagerState(
          db: db,
          table: table,
          createFilteringComposer: () => $$EventSnapshotsTableFilterComposer($db: db, $table: table),
          createOrderingComposer: () => $$EventSnapshotsTableOrderingComposer($db: db, $table: table),
          createComputedFieldComposer: () => $$EventSnapshotsTableAnnotationComposer($db: db, $table: table),
          updateCompanionCallback:
              ({
                Value<String> eventId = const Value.absent(),
                Value<String> version = const Value.absent(),
                Value<String> snapshotJson = const Value.absent(),
                Value<DateTime> syncedAt = const Value.absent(),
                Value<int> rowid = const Value.absent(),
              }) => EventSnapshotsCompanion(
                eventId: eventId,
                version: version,
                snapshotJson: snapshotJson,
                syncedAt: syncedAt,
                rowid: rowid,
              ),
          createCompanionCallback:
              ({
                required String eventId,
                required String version,
                required String snapshotJson,
                required DateTime syncedAt,
                Value<int> rowid = const Value.absent(),
              }) => EventSnapshotsCompanion.insert(
                eventId: eventId,
                version: version,
                snapshotJson: snapshotJson,
                syncedAt: syncedAt,
                rowid: rowid,
              ),
          withReferenceMapper: (p0) => p0
              .map(
                (e) => (
                  e.readTable<$EventSnapshotsTable, EventSnapshotRow>(table),
                  BaseReferences<_$AppDatabase, $EventSnapshotsTable, EventSnapshotRow>(db, table, e),
                ),
              )
              .toList(),
          prefetchHooksCallback: null,
        ),
      );
}

typedef $$EventSnapshotsTableProcessedTableManager =
    ProcessedTableManager<
      _$AppDatabase,
      $EventSnapshotsTable,
      EventSnapshotRow,
      $$EventSnapshotsTableFilterComposer,
      $$EventSnapshotsTableOrderingComposer,
      $$EventSnapshotsTableAnnotationComposer,
      $$EventSnapshotsTableCreateCompanionBuilder,
      $$EventSnapshotsTableUpdateCompanionBuilder,
      (EventSnapshotRow, BaseReferences<_$AppDatabase, $EventSnapshotsTable, EventSnapshotRow>),
      EventSnapshotRow,
      PrefetchHooks Function()
    >;
typedef $$ProductsTableCreateCompanionBuilder = ProductsCompanion Function({
  required String eventId,
  required String id,
  required String sku,
  required String name,
  Value<String?> brand,
  required double price,
  Value<String?> imageUrl,
  required String barcodes,
  required int remainingStock,
  Value<String> variantsJson,
  Value<int> rowid,
});
typedef $$ProductsTableUpdateCompanionBuilder = ProductsCompanion Function({
  Value<String> eventId,
  Value<String> id,
  Value<String> sku,
  Value<String> name,
  Value<String?> brand,
  Value<double> price,
  Value<String?> imageUrl,
  Value<String> barcodes,
  Value<int> remainingStock,
  Value<String> variantsJson,
  Value<int> rowid,
});

class $$ProductsTableFilterComposer extends Composer<_$AppDatabase, $ProductsTable> {
  $$ProductsTableFilterComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnFilters<String> get eventId =>
      $composableBuilder(column: $table.eventId, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get id => $composableBuilder(column: $table.id, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get sku => $composableBuilder(column: $table.sku, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get name => $composableBuilder(column: $table.name, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get brand =>
      $composableBuilder(column: $table.brand, builder: (column) => ColumnFilters(column));

  ColumnFilters<double> get price =>
      $composableBuilder(column: $table.price, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get imageUrl =>
      $composableBuilder(column: $table.imageUrl, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get barcodes =>
      $composableBuilder(column: $table.barcodes, builder: (column) => ColumnFilters(column));

  ColumnFilters<int> get remainingStock =>
      $composableBuilder(column: $table.remainingStock, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get variantsJson =>
      $composableBuilder(column: $table.variantsJson, builder: (column) => ColumnFilters(column));
}

class $$ProductsTableOrderingComposer extends Composer<_$AppDatabase, $ProductsTable> {
  $$ProductsTableOrderingComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnOrderings<String> get eventId =>
      $composableBuilder(column: $table.eventId, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get id => $composableBuilder(column: $table.id, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get sku =>
      $composableBuilder(column: $table.sku, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get name =>
      $composableBuilder(column: $table.name, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get brand =>
      $composableBuilder(column: $table.brand, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<double> get price =>
      $composableBuilder(column: $table.price, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get imageUrl =>
      $composableBuilder(column: $table.imageUrl, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get barcodes =>
      $composableBuilder(column: $table.barcodes, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<int> get remainingStock =>
      $composableBuilder(column: $table.remainingStock, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get variantsJson =>
      $composableBuilder(column: $table.variantsJson, builder: (column) => ColumnOrderings(column));
}

class $$ProductsTableAnnotationComposer extends Composer<_$AppDatabase, $ProductsTable> {
  $$ProductsTableAnnotationComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  GeneratedColumn<String> get eventId => $composableBuilder(column: $table.eventId, builder: (column) => column);

  GeneratedColumn<String> get id => $composableBuilder(column: $table.id, builder: (column) => column);

  GeneratedColumn<String> get sku => $composableBuilder(column: $table.sku, builder: (column) => column);

  GeneratedColumn<String> get name => $composableBuilder(column: $table.name, builder: (column) => column);

  GeneratedColumn<String> get brand => $composableBuilder(column: $table.brand, builder: (column) => column);

  GeneratedColumn<double> get price => $composableBuilder(column: $table.price, builder: (column) => column);

  GeneratedColumn<String> get imageUrl => $composableBuilder(column: $table.imageUrl, builder: (column) => column);

  GeneratedColumn<String> get barcodes => $composableBuilder(column: $table.barcodes, builder: (column) => column);

  GeneratedColumn<int> get remainingStock =>
      $composableBuilder(column: $table.remainingStock, builder: (column) => column);

  GeneratedColumn<String> get variantsJson =>
      $composableBuilder(column: $table.variantsJson, builder: (column) => column);
}

class $$ProductsTableTableManager
    extends
        RootTableManager<
          _$AppDatabase,
          $ProductsTable,
          ProductRow,
          $$ProductsTableFilterComposer,
          $$ProductsTableOrderingComposer,
          $$ProductsTableAnnotationComposer,
          $$ProductsTableCreateCompanionBuilder,
          $$ProductsTableUpdateCompanionBuilder,
          (ProductRow, BaseReferences<_$AppDatabase, $ProductsTable, ProductRow>),
          ProductRow,
          PrefetchHooks Function()
        > {
  $$ProductsTableTableManager(_$AppDatabase db, $ProductsTable table)
    : super(
        TableManagerState(
          db: db,
          table: table,
          createFilteringComposer: () => $$ProductsTableFilterComposer($db: db, $table: table),
          createOrderingComposer: () => $$ProductsTableOrderingComposer($db: db, $table: table),
          createComputedFieldComposer: () => $$ProductsTableAnnotationComposer($db: db, $table: table),
          updateCompanionCallback:
              ({
                Value<String> eventId = const Value.absent(),
                Value<String> id = const Value.absent(),
                Value<String> sku = const Value.absent(),
                Value<String> name = const Value.absent(),
                Value<String?> brand = const Value.absent(),
                Value<double> price = const Value.absent(),
                Value<String?> imageUrl = const Value.absent(),
                Value<String> barcodes = const Value.absent(),
                Value<int> remainingStock = const Value.absent(),
                Value<String> variantsJson = const Value.absent(),
                Value<int> rowid = const Value.absent(),
              }) => ProductsCompanion(
                eventId: eventId,
                id: id,
                sku: sku,
                name: name,
                brand: brand,
                price: price,
                imageUrl: imageUrl,
                barcodes: barcodes,
                remainingStock: remainingStock,
                variantsJson: variantsJson,
                rowid: rowid,
              ),
          createCompanionCallback:
              ({
                required String eventId,
                required String id,
                required String sku,
                required String name,
                Value<String?> brand = const Value.absent(),
                required double price,
                Value<String?> imageUrl = const Value.absent(),
                required String barcodes,
                required int remainingStock,
                Value<String> variantsJson = const Value.absent(),
                Value<int> rowid = const Value.absent(),
              }) => ProductsCompanion.insert(
                eventId: eventId,
                id: id,
                sku: sku,
                name: name,
                brand: brand,
                price: price,
                imageUrl: imageUrl,
                barcodes: barcodes,
                remainingStock: remainingStock,
                variantsJson: variantsJson,
                rowid: rowid,
              ),
          withReferenceMapper: (p0) => p0
              .map(
                (e) => (
                  e.readTable<$ProductsTable, ProductRow>(table),
                  BaseReferences<_$AppDatabase, $ProductsTable, ProductRow>(db, table, e),
                ),
              )
              .toList(),
          prefetchHooksCallback: null,
        ),
      );
}

typedef $$ProductsTableProcessedTableManager =
    ProcessedTableManager<
      _$AppDatabase,
      $ProductsTable,
      ProductRow,
      $$ProductsTableFilterComposer,
      $$ProductsTableOrderingComposer,
      $$ProductsTableAnnotationComposer,
      $$ProductsTableCreateCompanionBuilder,
      $$ProductsTableUpdateCompanionBuilder,
      (ProductRow, BaseReferences<_$AppDatabase, $ProductsTable, ProductRow>),
      ProductRow,
      PrefetchHooks Function()
    >;
typedef $$MembersTableCreateCompanionBuilder = MembersCompanion Function({
  required String qrCode,
  required String name,
  required String tier,
  Value<int> rowid,
});
typedef $$MembersTableUpdateCompanionBuilder = MembersCompanion Function({
  Value<String> qrCode,
  Value<String> name,
  Value<String> tier,
  Value<int> rowid,
});

class $$MembersTableFilterComposer extends Composer<_$AppDatabase, $MembersTable> {
  $$MembersTableFilterComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnFilters<String> get qrCode =>
      $composableBuilder(column: $table.qrCode, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get name => $composableBuilder(column: $table.name, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get tier => $composableBuilder(column: $table.tier, builder: (column) => ColumnFilters(column));
}

class $$MembersTableOrderingComposer extends Composer<_$AppDatabase, $MembersTable> {
  $$MembersTableOrderingComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnOrderings<String> get qrCode =>
      $composableBuilder(column: $table.qrCode, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get name =>
      $composableBuilder(column: $table.name, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get tier =>
      $composableBuilder(column: $table.tier, builder: (column) => ColumnOrderings(column));
}

class $$MembersTableAnnotationComposer extends Composer<_$AppDatabase, $MembersTable> {
  $$MembersTableAnnotationComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  GeneratedColumn<String> get qrCode => $composableBuilder(column: $table.qrCode, builder: (column) => column);

  GeneratedColumn<String> get name => $composableBuilder(column: $table.name, builder: (column) => column);

  GeneratedColumn<String> get tier => $composableBuilder(column: $table.tier, builder: (column) => column);
}

class $$MembersTableTableManager
    extends
        RootTableManager<
          _$AppDatabase,
          $MembersTable,
          MemberRow,
          $$MembersTableFilterComposer,
          $$MembersTableOrderingComposer,
          $$MembersTableAnnotationComposer,
          $$MembersTableCreateCompanionBuilder,
          $$MembersTableUpdateCompanionBuilder,
          (MemberRow, BaseReferences<_$AppDatabase, $MembersTable, MemberRow>),
          MemberRow,
          PrefetchHooks Function()
        > {
  $$MembersTableTableManager(_$AppDatabase db, $MembersTable table)
    : super(
        TableManagerState(
          db: db,
          table: table,
          createFilteringComposer: () => $$MembersTableFilterComposer($db: db, $table: table),
          createOrderingComposer: () => $$MembersTableOrderingComposer($db: db, $table: table),
          createComputedFieldComposer: () => $$MembersTableAnnotationComposer($db: db, $table: table),
          updateCompanionCallback: ({
            Value<String> qrCode = const Value.absent(),
            Value<String> name = const Value.absent(),
            Value<String> tier = const Value.absent(),
            Value<int> rowid = const Value.absent(),
          }) => MembersCompanion(qrCode: qrCode, name: name, tier: tier, rowid: rowid),
          createCompanionCallback: ({
            required String qrCode,
            required String name,
            required String tier,
            Value<int> rowid = const Value.absent(),
          }) => MembersCompanion.insert(qrCode: qrCode, name: name, tier: tier, rowid: rowid),
          withReferenceMapper: (p0) => p0
              .map(
                (e) => (
                  e.readTable<$MembersTable, MemberRow>(table),
                  BaseReferences<_$AppDatabase, $MembersTable, MemberRow>(db, table, e),
                ),
              )
              .toList(),
          prefetchHooksCallback: null,
        ),
      );
}

typedef $$MembersTableProcessedTableManager =
    ProcessedTableManager<
      _$AppDatabase,
      $MembersTable,
      MemberRow,
      $$MembersTableFilterComposer,
      $$MembersTableOrderingComposer,
      $$MembersTableAnnotationComposer,
      $$MembersTableCreateCompanionBuilder,
      $$MembersTableUpdateCompanionBuilder,
      (MemberRow, BaseReferences<_$AppDatabase, $MembersTable, MemberRow>),
      MemberRow,
      PrefetchHooks Function()
    >;
typedef $$PromotionsTableCreateCompanionBuilder = PromotionsCompanion Function({
  required String eventId,
  required String id,
  required String name,
  required String type,
  Value<double?> percentOff,
  Value<double?> amountOff,
  required String scope,
  required String productIds,
  Value<int?> minQty,
  Value<double?> minSubtotal,
  Value<String?> startsAt,
  Value<String?> endsAt,
  Value<int> priority,
  Value<int> rowid,
});
typedef $$PromotionsTableUpdateCompanionBuilder = PromotionsCompanion Function({
  Value<String> eventId,
  Value<String> id,
  Value<String> name,
  Value<String> type,
  Value<double?> percentOff,
  Value<double?> amountOff,
  Value<String> scope,
  Value<String> productIds,
  Value<int?> minQty,
  Value<double?> minSubtotal,
  Value<String?> startsAt,
  Value<String?> endsAt,
  Value<int> priority,
  Value<int> rowid,
});

class $$PromotionsTableFilterComposer extends Composer<_$AppDatabase, $PromotionsTable> {
  $$PromotionsTableFilterComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnFilters<String> get eventId =>
      $composableBuilder(column: $table.eventId, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get id => $composableBuilder(column: $table.id, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get name => $composableBuilder(column: $table.name, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get type => $composableBuilder(column: $table.type, builder: (column) => ColumnFilters(column));

  ColumnFilters<double> get percentOff =>
      $composableBuilder(column: $table.percentOff, builder: (column) => ColumnFilters(column));

  ColumnFilters<double> get amountOff =>
      $composableBuilder(column: $table.amountOff, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get scope =>
      $composableBuilder(column: $table.scope, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get productIds =>
      $composableBuilder(column: $table.productIds, builder: (column) => ColumnFilters(column));

  ColumnFilters<int> get minQty =>
      $composableBuilder(column: $table.minQty, builder: (column) => ColumnFilters(column));

  ColumnFilters<double> get minSubtotal =>
      $composableBuilder(column: $table.minSubtotal, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get startsAt =>
      $composableBuilder(column: $table.startsAt, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get endsAt =>
      $composableBuilder(column: $table.endsAt, builder: (column) => ColumnFilters(column));

  ColumnFilters<int> get priority =>
      $composableBuilder(column: $table.priority, builder: (column) => ColumnFilters(column));
}

class $$PromotionsTableOrderingComposer extends Composer<_$AppDatabase, $PromotionsTable> {
  $$PromotionsTableOrderingComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnOrderings<String> get eventId =>
      $composableBuilder(column: $table.eventId, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get id => $composableBuilder(column: $table.id, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get name =>
      $composableBuilder(column: $table.name, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get type =>
      $composableBuilder(column: $table.type, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<double> get percentOff =>
      $composableBuilder(column: $table.percentOff, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<double> get amountOff =>
      $composableBuilder(column: $table.amountOff, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get scope =>
      $composableBuilder(column: $table.scope, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get productIds =>
      $composableBuilder(column: $table.productIds, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<int> get minQty =>
      $composableBuilder(column: $table.minQty, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<double> get minSubtotal =>
      $composableBuilder(column: $table.minSubtotal, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get startsAt =>
      $composableBuilder(column: $table.startsAt, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get endsAt =>
      $composableBuilder(column: $table.endsAt, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<int> get priority =>
      $composableBuilder(column: $table.priority, builder: (column) => ColumnOrderings(column));
}

class $$PromotionsTableAnnotationComposer extends Composer<_$AppDatabase, $PromotionsTable> {
  $$PromotionsTableAnnotationComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  GeneratedColumn<String> get eventId => $composableBuilder(column: $table.eventId, builder: (column) => column);

  GeneratedColumn<String> get id => $composableBuilder(column: $table.id, builder: (column) => column);

  GeneratedColumn<String> get name => $composableBuilder(column: $table.name, builder: (column) => column);

  GeneratedColumn<String> get type => $composableBuilder(column: $table.type, builder: (column) => column);

  GeneratedColumn<double> get percentOff => $composableBuilder(column: $table.percentOff, builder: (column) => column);

  GeneratedColumn<double> get amountOff => $composableBuilder(column: $table.amountOff, builder: (column) => column);

  GeneratedColumn<String> get scope => $composableBuilder(column: $table.scope, builder: (column) => column);

  GeneratedColumn<String> get productIds => $composableBuilder(column: $table.productIds, builder: (column) => column);

  GeneratedColumn<int> get minQty => $composableBuilder(column: $table.minQty, builder: (column) => column);

  GeneratedColumn<double> get minSubtotal =>
      $composableBuilder(column: $table.minSubtotal, builder: (column) => column);

  GeneratedColumn<String> get startsAt => $composableBuilder(column: $table.startsAt, builder: (column) => column);

  GeneratedColumn<String> get endsAt => $composableBuilder(column: $table.endsAt, builder: (column) => column);

  GeneratedColumn<int> get priority => $composableBuilder(column: $table.priority, builder: (column) => column);
}

class $$PromotionsTableTableManager
    extends
        RootTableManager<
          _$AppDatabase,
          $PromotionsTable,
          PromotionRow,
          $$PromotionsTableFilterComposer,
          $$PromotionsTableOrderingComposer,
          $$PromotionsTableAnnotationComposer,
          $$PromotionsTableCreateCompanionBuilder,
          $$PromotionsTableUpdateCompanionBuilder,
          (PromotionRow, BaseReferences<_$AppDatabase, $PromotionsTable, PromotionRow>),
          PromotionRow,
          PrefetchHooks Function()
        > {
  $$PromotionsTableTableManager(_$AppDatabase db, $PromotionsTable table)
    : super(
        TableManagerState(
          db: db,
          table: table,
          createFilteringComposer: () => $$PromotionsTableFilterComposer($db: db, $table: table),
          createOrderingComposer: () => $$PromotionsTableOrderingComposer($db: db, $table: table),
          createComputedFieldComposer: () => $$PromotionsTableAnnotationComposer($db: db, $table: table),
          updateCompanionCallback:
              ({
                Value<String> eventId = const Value.absent(),
                Value<String> id = const Value.absent(),
                Value<String> name = const Value.absent(),
                Value<String> type = const Value.absent(),
                Value<double?> percentOff = const Value.absent(),
                Value<double?> amountOff = const Value.absent(),
                Value<String> scope = const Value.absent(),
                Value<String> productIds = const Value.absent(),
                Value<int?> minQty = const Value.absent(),
                Value<double?> minSubtotal = const Value.absent(),
                Value<String?> startsAt = const Value.absent(),
                Value<String?> endsAt = const Value.absent(),
                Value<int> priority = const Value.absent(),
                Value<int> rowid = const Value.absent(),
              }) => PromotionsCompanion(
                eventId: eventId,
                id: id,
                name: name,
                type: type,
                percentOff: percentOff,
                amountOff: amountOff,
                scope: scope,
                productIds: productIds,
                minQty: minQty,
                minSubtotal: minSubtotal,
                startsAt: startsAt,
                endsAt: endsAt,
                priority: priority,
                rowid: rowid,
              ),
          createCompanionCallback:
              ({
                required String eventId,
                required String id,
                required String name,
                required String type,
                Value<double?> percentOff = const Value.absent(),
                Value<double?> amountOff = const Value.absent(),
                required String scope,
                required String productIds,
                Value<int?> minQty = const Value.absent(),
                Value<double?> minSubtotal = const Value.absent(),
                Value<String?> startsAt = const Value.absent(),
                Value<String?> endsAt = const Value.absent(),
                Value<int> priority = const Value.absent(),
                Value<int> rowid = const Value.absent(),
              }) => PromotionsCompanion.insert(
                eventId: eventId,
                id: id,
                name: name,
                type: type,
                percentOff: percentOff,
                amountOff: amountOff,
                scope: scope,
                productIds: productIds,
                minQty: minQty,
                minSubtotal: minSubtotal,
                startsAt: startsAt,
                endsAt: endsAt,
                priority: priority,
                rowid: rowid,
              ),
          withReferenceMapper: (p0) => p0
              .map(
                (e) => (
                  e.readTable<$PromotionsTable, PromotionRow>(table),
                  BaseReferences<_$AppDatabase, $PromotionsTable, PromotionRow>(db, table, e),
                ),
              )
              .toList(),
          prefetchHooksCallback: null,
        ),
      );
}

typedef $$PromotionsTableProcessedTableManager =
    ProcessedTableManager<
      _$AppDatabase,
      $PromotionsTable,
      PromotionRow,
      $$PromotionsTableFilterComposer,
      $$PromotionsTableOrderingComposer,
      $$PromotionsTableAnnotationComposer,
      $$PromotionsTableCreateCompanionBuilder,
      $$PromotionsTableUpdateCompanionBuilder,
      (PromotionRow, BaseReferences<_$AppDatabase, $PromotionsTable, PromotionRow>),
      PromotionRow,
      PrefetchHooks Function()
    >;
typedef $$PendingOrdersTableCreateCompanionBuilder = PendingOrdersCompanion Function({
  required String idempotencyKey,
  required String eventId,
  required String customerName,
  Value<String?> customerQr,
  Value<String?> customerTier,
  required String linesJson,
  required String status,
  required DateTime createdAt,
  Value<String?> orderNumberDisplay,
  Value<String?> referenceId,
  Value<double?> total,
  Value<String?> errorMessage,
  Value<int?> orderNumber,
  Value<String?> serverOrderId,
  Value<int> rowid,
});
typedef $$PendingOrdersTableUpdateCompanionBuilder = PendingOrdersCompanion Function({
  Value<String> idempotencyKey,
  Value<String> eventId,
  Value<String> customerName,
  Value<String?> customerQr,
  Value<String?> customerTier,
  Value<String> linesJson,
  Value<String> status,
  Value<DateTime> createdAt,
  Value<String?> orderNumberDisplay,
  Value<String?> referenceId,
  Value<double?> total,
  Value<String?> errorMessage,
  Value<int?> orderNumber,
  Value<String?> serverOrderId,
  Value<int> rowid,
});

class $$PendingOrdersTableFilterComposer extends Composer<_$AppDatabase, $PendingOrdersTable> {
  $$PendingOrdersTableFilterComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnFilters<String> get idempotencyKey =>
      $composableBuilder(column: $table.idempotencyKey, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get eventId =>
      $composableBuilder(column: $table.eventId, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get customerName =>
      $composableBuilder(column: $table.customerName, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get customerQr =>
      $composableBuilder(column: $table.customerQr, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get customerTier =>
      $composableBuilder(column: $table.customerTier, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get linesJson =>
      $composableBuilder(column: $table.linesJson, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get status =>
      $composableBuilder(column: $table.status, builder: (column) => ColumnFilters(column));

  ColumnFilters<DateTime> get createdAt =>
      $composableBuilder(column: $table.createdAt, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get orderNumberDisplay =>
      $composableBuilder(column: $table.orderNumberDisplay, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get referenceId =>
      $composableBuilder(column: $table.referenceId, builder: (column) => ColumnFilters(column));

  ColumnFilters<double> get total =>
      $composableBuilder(column: $table.total, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get errorMessage =>
      $composableBuilder(column: $table.errorMessage, builder: (column) => ColumnFilters(column));

  ColumnFilters<int> get orderNumber =>
      $composableBuilder(column: $table.orderNumber, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get serverOrderId =>
      $composableBuilder(column: $table.serverOrderId, builder: (column) => ColumnFilters(column));
}

class $$PendingOrdersTableOrderingComposer extends Composer<_$AppDatabase, $PendingOrdersTable> {
  $$PendingOrdersTableOrderingComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnOrderings<String> get idempotencyKey =>
      $composableBuilder(column: $table.idempotencyKey, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get eventId =>
      $composableBuilder(column: $table.eventId, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get customerName =>
      $composableBuilder(column: $table.customerName, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get customerQr =>
      $composableBuilder(column: $table.customerQr, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get customerTier =>
      $composableBuilder(column: $table.customerTier, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get linesJson =>
      $composableBuilder(column: $table.linesJson, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get status =>
      $composableBuilder(column: $table.status, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<DateTime> get createdAt =>
      $composableBuilder(column: $table.createdAt, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get orderNumberDisplay =>
      $composableBuilder(column: $table.orderNumberDisplay, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get referenceId =>
      $composableBuilder(column: $table.referenceId, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<double> get total =>
      $composableBuilder(column: $table.total, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get errorMessage =>
      $composableBuilder(column: $table.errorMessage, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<int> get orderNumber =>
      $composableBuilder(column: $table.orderNumber, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get serverOrderId =>
      $composableBuilder(column: $table.serverOrderId, builder: (column) => ColumnOrderings(column));
}

class $$PendingOrdersTableAnnotationComposer extends Composer<_$AppDatabase, $PendingOrdersTable> {
  $$PendingOrdersTableAnnotationComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  GeneratedColumn<String> get idempotencyKey =>
      $composableBuilder(column: $table.idempotencyKey, builder: (column) => column);

  GeneratedColumn<String> get eventId => $composableBuilder(column: $table.eventId, builder: (column) => column);

  GeneratedColumn<String> get customerName =>
      $composableBuilder(column: $table.customerName, builder: (column) => column);

  GeneratedColumn<String> get customerQr => $composableBuilder(column: $table.customerQr, builder: (column) => column);

  GeneratedColumn<String> get customerTier =>
      $composableBuilder(column: $table.customerTier, builder: (column) => column);

  GeneratedColumn<String> get linesJson => $composableBuilder(column: $table.linesJson, builder: (column) => column);

  GeneratedColumn<String> get status => $composableBuilder(column: $table.status, builder: (column) => column);

  GeneratedColumn<DateTime> get createdAt => $composableBuilder(column: $table.createdAt, builder: (column) => column);

  GeneratedColumn<String> get orderNumberDisplay =>
      $composableBuilder(column: $table.orderNumberDisplay, builder: (column) => column);

  GeneratedColumn<String> get referenceId =>
      $composableBuilder(column: $table.referenceId, builder: (column) => column);

  GeneratedColumn<double> get total => $composableBuilder(column: $table.total, builder: (column) => column);

  GeneratedColumn<String> get errorMessage =>
      $composableBuilder(column: $table.errorMessage, builder: (column) => column);

  GeneratedColumn<int> get orderNumber => $composableBuilder(column: $table.orderNumber, builder: (column) => column);

  GeneratedColumn<String> get serverOrderId =>
      $composableBuilder(column: $table.serverOrderId, builder: (column) => column);
}

class $$PendingOrdersTableTableManager
    extends
        RootTableManager<
          _$AppDatabase,
          $PendingOrdersTable,
          PendingOrderRow,
          $$PendingOrdersTableFilterComposer,
          $$PendingOrdersTableOrderingComposer,
          $$PendingOrdersTableAnnotationComposer,
          $$PendingOrdersTableCreateCompanionBuilder,
          $$PendingOrdersTableUpdateCompanionBuilder,
          (PendingOrderRow, BaseReferences<_$AppDatabase, $PendingOrdersTable, PendingOrderRow>),
          PendingOrderRow,
          PrefetchHooks Function()
        > {
  $$PendingOrdersTableTableManager(_$AppDatabase db, $PendingOrdersTable table)
    : super(
        TableManagerState(
          db: db,
          table: table,
          createFilteringComposer: () => $$PendingOrdersTableFilterComposer($db: db, $table: table),
          createOrderingComposer: () => $$PendingOrdersTableOrderingComposer($db: db, $table: table),
          createComputedFieldComposer: () => $$PendingOrdersTableAnnotationComposer($db: db, $table: table),
          updateCompanionCallback:
              ({
                Value<String> idempotencyKey = const Value.absent(),
                Value<String> eventId = const Value.absent(),
                Value<String> customerName = const Value.absent(),
                Value<String?> customerQr = const Value.absent(),
                Value<String?> customerTier = const Value.absent(),
                Value<String> linesJson = const Value.absent(),
                Value<String> status = const Value.absent(),
                Value<DateTime> createdAt = const Value.absent(),
                Value<String?> orderNumberDisplay = const Value.absent(),
                Value<String?> referenceId = const Value.absent(),
                Value<double?> total = const Value.absent(),
                Value<String?> errorMessage = const Value.absent(),
                Value<int?> orderNumber = const Value.absent(),
                Value<String?> serverOrderId = const Value.absent(),
                Value<int> rowid = const Value.absent(),
              }) => PendingOrdersCompanion(
                idempotencyKey: idempotencyKey,
                eventId: eventId,
                customerName: customerName,
                customerQr: customerQr,
                customerTier: customerTier,
                linesJson: linesJson,
                status: status,
                createdAt: createdAt,
                orderNumberDisplay: orderNumberDisplay,
                referenceId: referenceId,
                total: total,
                errorMessage: errorMessage,
                orderNumber: orderNumber,
                serverOrderId: serverOrderId,
                rowid: rowid,
              ),
          createCompanionCallback:
              ({
                required String idempotencyKey,
                required String eventId,
                required String customerName,
                Value<String?> customerQr = const Value.absent(),
                Value<String?> customerTier = const Value.absent(),
                required String linesJson,
                required String status,
                required DateTime createdAt,
                Value<String?> orderNumberDisplay = const Value.absent(),
                Value<String?> referenceId = const Value.absent(),
                Value<double?> total = const Value.absent(),
                Value<String?> errorMessage = const Value.absent(),
                Value<int?> orderNumber = const Value.absent(),
                Value<String?> serverOrderId = const Value.absent(),
                Value<int> rowid = const Value.absent(),
              }) => PendingOrdersCompanion.insert(
                idempotencyKey: idempotencyKey,
                eventId: eventId,
                customerName: customerName,
                customerQr: customerQr,
                customerTier: customerTier,
                linesJson: linesJson,
                status: status,
                createdAt: createdAt,
                orderNumberDisplay: orderNumberDisplay,
                referenceId: referenceId,
                total: total,
                errorMessage: errorMessage,
                orderNumber: orderNumber,
                serverOrderId: serverOrderId,
                rowid: rowid,
              ),
          withReferenceMapper: (p0) => p0
              .map(
                (e) => (
                  e.readTable<$PendingOrdersTable, PendingOrderRow>(table),
                  BaseReferences<_$AppDatabase, $PendingOrdersTable, PendingOrderRow>(db, table, e),
                ),
              )
              .toList(),
          prefetchHooksCallback: null,
        ),
      );
}

typedef $$PendingOrdersTableProcessedTableManager =
    ProcessedTableManager<
      _$AppDatabase,
      $PendingOrdersTable,
      PendingOrderRow,
      $$PendingOrdersTableFilterComposer,
      $$PendingOrdersTableOrderingComposer,
      $$PendingOrdersTableAnnotationComposer,
      $$PendingOrdersTableCreateCompanionBuilder,
      $$PendingOrdersTableUpdateCompanionBuilder,
      (PendingOrderRow, BaseReferences<_$AppDatabase, $PendingOrdersTable, PendingOrderRow>),
      PendingOrderRow,
      PrefetchHooks Function()
    >;
typedef $$ReceiptsTableCreateCompanionBuilder = ReceiptsCompanion Function({
  required String orderKey,
  required String eventId,
  required String orderNumberDisplay,
  required String receiptText,
  required bool printedOk,
  required DateTime createdAt,
  Value<String?> backendOrderId,
  Value<int> rowid,
});
typedef $$ReceiptsTableUpdateCompanionBuilder = ReceiptsCompanion Function({
  Value<String> orderKey,
  Value<String> eventId,
  Value<String> orderNumberDisplay,
  Value<String> receiptText,
  Value<bool> printedOk,
  Value<DateTime> createdAt,
  Value<String?> backendOrderId,
  Value<int> rowid,
});

class $$ReceiptsTableFilterComposer extends Composer<_$AppDatabase, $ReceiptsTable> {
  $$ReceiptsTableFilterComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnFilters<String> get orderKey =>
      $composableBuilder(column: $table.orderKey, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get eventId =>
      $composableBuilder(column: $table.eventId, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get orderNumberDisplay =>
      $composableBuilder(column: $table.orderNumberDisplay, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get receiptText =>
      $composableBuilder(column: $table.receiptText, builder: (column) => ColumnFilters(column));

  ColumnFilters<bool> get printedOk =>
      $composableBuilder(column: $table.printedOk, builder: (column) => ColumnFilters(column));

  ColumnFilters<DateTime> get createdAt =>
      $composableBuilder(column: $table.createdAt, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get backendOrderId =>
      $composableBuilder(column: $table.backendOrderId, builder: (column) => ColumnFilters(column));
}

class $$ReceiptsTableOrderingComposer extends Composer<_$AppDatabase, $ReceiptsTable> {
  $$ReceiptsTableOrderingComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnOrderings<String> get orderKey =>
      $composableBuilder(column: $table.orderKey, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get eventId =>
      $composableBuilder(column: $table.eventId, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get orderNumberDisplay =>
      $composableBuilder(column: $table.orderNumberDisplay, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get receiptText =>
      $composableBuilder(column: $table.receiptText, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<bool> get printedOk =>
      $composableBuilder(column: $table.printedOk, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<DateTime> get createdAt =>
      $composableBuilder(column: $table.createdAt, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get backendOrderId =>
      $composableBuilder(column: $table.backendOrderId, builder: (column) => ColumnOrderings(column));
}

class $$ReceiptsTableAnnotationComposer extends Composer<_$AppDatabase, $ReceiptsTable> {
  $$ReceiptsTableAnnotationComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  GeneratedColumn<String> get orderKey => $composableBuilder(column: $table.orderKey, builder: (column) => column);

  GeneratedColumn<String> get eventId => $composableBuilder(column: $table.eventId, builder: (column) => column);

  GeneratedColumn<String> get orderNumberDisplay =>
      $composableBuilder(column: $table.orderNumberDisplay, builder: (column) => column);

  GeneratedColumn<String> get receiptText =>
      $composableBuilder(column: $table.receiptText, builder: (column) => column);

  GeneratedColumn<bool> get printedOk => $composableBuilder(column: $table.printedOk, builder: (column) => column);

  GeneratedColumn<DateTime> get createdAt => $composableBuilder(column: $table.createdAt, builder: (column) => column);

  GeneratedColumn<String> get backendOrderId =>
      $composableBuilder(column: $table.backendOrderId, builder: (column) => column);
}

class $$ReceiptsTableTableManager
    extends
        RootTableManager<
          _$AppDatabase,
          $ReceiptsTable,
          ReceiptRow,
          $$ReceiptsTableFilterComposer,
          $$ReceiptsTableOrderingComposer,
          $$ReceiptsTableAnnotationComposer,
          $$ReceiptsTableCreateCompanionBuilder,
          $$ReceiptsTableUpdateCompanionBuilder,
          (ReceiptRow, BaseReferences<_$AppDatabase, $ReceiptsTable, ReceiptRow>),
          ReceiptRow,
          PrefetchHooks Function()
        > {
  $$ReceiptsTableTableManager(_$AppDatabase db, $ReceiptsTable table)
    : super(
        TableManagerState(
          db: db,
          table: table,
          createFilteringComposer: () => $$ReceiptsTableFilterComposer($db: db, $table: table),
          createOrderingComposer: () => $$ReceiptsTableOrderingComposer($db: db, $table: table),
          createComputedFieldComposer: () => $$ReceiptsTableAnnotationComposer($db: db, $table: table),
          updateCompanionCallback:
              ({
                Value<String> orderKey = const Value.absent(),
                Value<String> eventId = const Value.absent(),
                Value<String> orderNumberDisplay = const Value.absent(),
                Value<String> receiptText = const Value.absent(),
                Value<bool> printedOk = const Value.absent(),
                Value<DateTime> createdAt = const Value.absent(),
                Value<String?> backendOrderId = const Value.absent(),
                Value<int> rowid = const Value.absent(),
              }) => ReceiptsCompanion(
                orderKey: orderKey,
                eventId: eventId,
                orderNumberDisplay: orderNumberDisplay,
                receiptText: receiptText,
                printedOk: printedOk,
                createdAt: createdAt,
                backendOrderId: backendOrderId,
                rowid: rowid,
              ),
          createCompanionCallback:
              ({
                required String orderKey,
                required String eventId,
                required String orderNumberDisplay,
                required String receiptText,
                required bool printedOk,
                required DateTime createdAt,
                Value<String?> backendOrderId = const Value.absent(),
                Value<int> rowid = const Value.absent(),
              }) => ReceiptsCompanion.insert(
                orderKey: orderKey,
                eventId: eventId,
                orderNumberDisplay: orderNumberDisplay,
                receiptText: receiptText,
                printedOk: printedOk,
                createdAt: createdAt,
                backendOrderId: backendOrderId,
                rowid: rowid,
              ),
          withReferenceMapper: (p0) => p0
              .map(
                (e) => (
                  e.readTable<$ReceiptsTable, ReceiptRow>(table),
                  BaseReferences<_$AppDatabase, $ReceiptsTable, ReceiptRow>(db, table, e),
                ),
              )
              .toList(),
          prefetchHooksCallback: null,
        ),
      );
}

typedef $$ReceiptsTableProcessedTableManager =
    ProcessedTableManager<
      _$AppDatabase,
      $ReceiptsTable,
      ReceiptRow,
      $$ReceiptsTableFilterComposer,
      $$ReceiptsTableOrderingComposer,
      $$ReceiptsTableAnnotationComposer,
      $$ReceiptsTableCreateCompanionBuilder,
      $$ReceiptsTableUpdateCompanionBuilder,
      (ReceiptRow, BaseReferences<_$AppDatabase, $ReceiptsTable, ReceiptRow>),
      ReceiptRow,
      PrefetchHooks Function()
    >;
typedef $$CartDraftsTableCreateCompanionBuilder = CartDraftsCompanion Function({
  required String id,
  required String eventId,
  required String label,
  Value<String?> customerJson,
  required String linesJson,
  required DateTime createdAt,
  Value<bool> live,
  Value<int> rowid,
});
typedef $$CartDraftsTableUpdateCompanionBuilder = CartDraftsCompanion Function({
  Value<String> id,
  Value<String> eventId,
  Value<String> label,
  Value<String?> customerJson,
  Value<String> linesJson,
  Value<DateTime> createdAt,
  Value<bool> live,
  Value<int> rowid,
});

class $$CartDraftsTableFilterComposer extends Composer<_$AppDatabase, $CartDraftsTable> {
  $$CartDraftsTableFilterComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnFilters<String> get id => $composableBuilder(column: $table.id, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get eventId =>
      $composableBuilder(column: $table.eventId, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get label =>
      $composableBuilder(column: $table.label, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get customerJson =>
      $composableBuilder(column: $table.customerJson, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get linesJson =>
      $composableBuilder(column: $table.linesJson, builder: (column) => ColumnFilters(column));

  ColumnFilters<DateTime> get createdAt =>
      $composableBuilder(column: $table.createdAt, builder: (column) => ColumnFilters(column));

  ColumnFilters<bool> get live => $composableBuilder(column: $table.live, builder: (column) => ColumnFilters(column));
}

class $$CartDraftsTableOrderingComposer extends Composer<_$AppDatabase, $CartDraftsTable> {
  $$CartDraftsTableOrderingComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnOrderings<String> get id => $composableBuilder(column: $table.id, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get eventId =>
      $composableBuilder(column: $table.eventId, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get label =>
      $composableBuilder(column: $table.label, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get customerJson =>
      $composableBuilder(column: $table.customerJson, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get linesJson =>
      $composableBuilder(column: $table.linesJson, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<DateTime> get createdAt =>
      $composableBuilder(column: $table.createdAt, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<bool> get live =>
      $composableBuilder(column: $table.live, builder: (column) => ColumnOrderings(column));
}

class $$CartDraftsTableAnnotationComposer extends Composer<_$AppDatabase, $CartDraftsTable> {
  $$CartDraftsTableAnnotationComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  GeneratedColumn<String> get id => $composableBuilder(column: $table.id, builder: (column) => column);

  GeneratedColumn<String> get eventId => $composableBuilder(column: $table.eventId, builder: (column) => column);

  GeneratedColumn<String> get label => $composableBuilder(column: $table.label, builder: (column) => column);

  GeneratedColumn<String> get customerJson =>
      $composableBuilder(column: $table.customerJson, builder: (column) => column);

  GeneratedColumn<String> get linesJson => $composableBuilder(column: $table.linesJson, builder: (column) => column);

  GeneratedColumn<DateTime> get createdAt => $composableBuilder(column: $table.createdAt, builder: (column) => column);

  GeneratedColumn<bool> get live => $composableBuilder(column: $table.live, builder: (column) => column);
}

class $$CartDraftsTableTableManager
    extends
        RootTableManager<
          _$AppDatabase,
          $CartDraftsTable,
          CartDraftRow,
          $$CartDraftsTableFilterComposer,
          $$CartDraftsTableOrderingComposer,
          $$CartDraftsTableAnnotationComposer,
          $$CartDraftsTableCreateCompanionBuilder,
          $$CartDraftsTableUpdateCompanionBuilder,
          (CartDraftRow, BaseReferences<_$AppDatabase, $CartDraftsTable, CartDraftRow>),
          CartDraftRow,
          PrefetchHooks Function()
        > {
  $$CartDraftsTableTableManager(_$AppDatabase db, $CartDraftsTable table)
    : super(
        TableManagerState(
          db: db,
          table: table,
          createFilteringComposer: () => $$CartDraftsTableFilterComposer($db: db, $table: table),
          createOrderingComposer: () => $$CartDraftsTableOrderingComposer($db: db, $table: table),
          createComputedFieldComposer: () => $$CartDraftsTableAnnotationComposer($db: db, $table: table),
          updateCompanionCallback:
              ({
                Value<String> id = const Value.absent(),
                Value<String> eventId = const Value.absent(),
                Value<String> label = const Value.absent(),
                Value<String?> customerJson = const Value.absent(),
                Value<String> linesJson = const Value.absent(),
                Value<DateTime> createdAt = const Value.absent(),
                Value<bool> live = const Value.absent(),
                Value<int> rowid = const Value.absent(),
              }) => CartDraftsCompanion(
                id: id,
                eventId: eventId,
                label: label,
                customerJson: customerJson,
                linesJson: linesJson,
                createdAt: createdAt,
                live: live,
                rowid: rowid,
              ),
          createCompanionCallback:
              ({
                required String id,
                required String eventId,
                required String label,
                Value<String?> customerJson = const Value.absent(),
                required String linesJson,
                required DateTime createdAt,
                Value<bool> live = const Value.absent(),
                Value<int> rowid = const Value.absent(),
              }) => CartDraftsCompanion.insert(
                id: id,
                eventId: eventId,
                label: label,
                customerJson: customerJson,
                linesJson: linesJson,
                createdAt: createdAt,
                live: live,
                rowid: rowid,
              ),
          withReferenceMapper: (p0) => p0
              .map(
                (e) => (
                  e.readTable<$CartDraftsTable, CartDraftRow>(table),
                  BaseReferences<_$AppDatabase, $CartDraftsTable, CartDraftRow>(db, table, e),
                ),
              )
              .toList(),
          prefetchHooksCallback: null,
        ),
      );
}

typedef $$CartDraftsTableProcessedTableManager =
    ProcessedTableManager<
      _$AppDatabase,
      $CartDraftsTable,
      CartDraftRow,
      $$CartDraftsTableFilterComposer,
      $$CartDraftsTableOrderingComposer,
      $$CartDraftsTableAnnotationComposer,
      $$CartDraftsTableCreateCompanionBuilder,
      $$CartDraftsTableUpdateCompanionBuilder,
      (CartDraftRow, BaseReferences<_$AppDatabase, $CartDraftsTable, CartDraftRow>),
      CartDraftRow,
      PrefetchHooks Function()
    >;
typedef $$NumberBlocksTableCreateCompanionBuilder = NumberBlocksCompanion Function({
  required String eventId,
  required String businessDate,
  required int rangeStart,
  required int rangeEnd,
  required int nextUnused,
  Value<int> rowid,
});
typedef $$NumberBlocksTableUpdateCompanionBuilder = NumberBlocksCompanion Function({
  Value<String> eventId,
  Value<String> businessDate,
  Value<int> rangeStart,
  Value<int> rangeEnd,
  Value<int> nextUnused,
  Value<int> rowid,
});

class $$NumberBlocksTableFilterComposer extends Composer<_$AppDatabase, $NumberBlocksTable> {
  $$NumberBlocksTableFilterComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnFilters<String> get eventId =>
      $composableBuilder(column: $table.eventId, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get businessDate =>
      $composableBuilder(column: $table.businessDate, builder: (column) => ColumnFilters(column));

  ColumnFilters<int> get rangeStart =>
      $composableBuilder(column: $table.rangeStart, builder: (column) => ColumnFilters(column));

  ColumnFilters<int> get rangeEnd =>
      $composableBuilder(column: $table.rangeEnd, builder: (column) => ColumnFilters(column));

  ColumnFilters<int> get nextUnused =>
      $composableBuilder(column: $table.nextUnused, builder: (column) => ColumnFilters(column));
}

class $$NumberBlocksTableOrderingComposer extends Composer<_$AppDatabase, $NumberBlocksTable> {
  $$NumberBlocksTableOrderingComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnOrderings<String> get eventId =>
      $composableBuilder(column: $table.eventId, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get businessDate =>
      $composableBuilder(column: $table.businessDate, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<int> get rangeStart =>
      $composableBuilder(column: $table.rangeStart, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<int> get rangeEnd =>
      $composableBuilder(column: $table.rangeEnd, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<int> get nextUnused =>
      $composableBuilder(column: $table.nextUnused, builder: (column) => ColumnOrderings(column));
}

class $$NumberBlocksTableAnnotationComposer extends Composer<_$AppDatabase, $NumberBlocksTable> {
  $$NumberBlocksTableAnnotationComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  GeneratedColumn<String> get eventId => $composableBuilder(column: $table.eventId, builder: (column) => column);

  GeneratedColumn<String> get businessDate =>
      $composableBuilder(column: $table.businessDate, builder: (column) => column);

  GeneratedColumn<int> get rangeStart => $composableBuilder(column: $table.rangeStart, builder: (column) => column);

  GeneratedColumn<int> get rangeEnd => $composableBuilder(column: $table.rangeEnd, builder: (column) => column);

  GeneratedColumn<int> get nextUnused => $composableBuilder(column: $table.nextUnused, builder: (column) => column);
}

class $$NumberBlocksTableTableManager
    extends
        RootTableManager<
          _$AppDatabase,
          $NumberBlocksTable,
          NumberBlockRow,
          $$NumberBlocksTableFilterComposer,
          $$NumberBlocksTableOrderingComposer,
          $$NumberBlocksTableAnnotationComposer,
          $$NumberBlocksTableCreateCompanionBuilder,
          $$NumberBlocksTableUpdateCompanionBuilder,
          (NumberBlockRow, BaseReferences<_$AppDatabase, $NumberBlocksTable, NumberBlockRow>),
          NumberBlockRow,
          PrefetchHooks Function()
        > {
  $$NumberBlocksTableTableManager(_$AppDatabase db, $NumberBlocksTable table)
    : super(
        TableManagerState(
          db: db,
          table: table,
          createFilteringComposer: () => $$NumberBlocksTableFilterComposer($db: db, $table: table),
          createOrderingComposer: () => $$NumberBlocksTableOrderingComposer($db: db, $table: table),
          createComputedFieldComposer: () => $$NumberBlocksTableAnnotationComposer($db: db, $table: table),
          updateCompanionCallback:
              ({
                Value<String> eventId = const Value.absent(),
                Value<String> businessDate = const Value.absent(),
                Value<int> rangeStart = const Value.absent(),
                Value<int> rangeEnd = const Value.absent(),
                Value<int> nextUnused = const Value.absent(),
                Value<int> rowid = const Value.absent(),
              }) => NumberBlocksCompanion(
                eventId: eventId,
                businessDate: businessDate,
                rangeStart: rangeStart,
                rangeEnd: rangeEnd,
                nextUnused: nextUnused,
                rowid: rowid,
              ),
          createCompanionCallback:
              ({
                required String eventId,
                required String businessDate,
                required int rangeStart,
                required int rangeEnd,
                required int nextUnused,
                Value<int> rowid = const Value.absent(),
              }) => NumberBlocksCompanion.insert(
                eventId: eventId,
                businessDate: businessDate,
                rangeStart: rangeStart,
                rangeEnd: rangeEnd,
                nextUnused: nextUnused,
                rowid: rowid,
              ),
          withReferenceMapper: (p0) => p0
              .map(
                (e) => (
                  e.readTable<$NumberBlocksTable, NumberBlockRow>(table),
                  BaseReferences<_$AppDatabase, $NumberBlocksTable, NumberBlockRow>(db, table, e),
                ),
              )
              .toList(),
          prefetchHooksCallback: null,
        ),
      );
}

typedef $$NumberBlocksTableProcessedTableManager =
    ProcessedTableManager<
      _$AppDatabase,
      $NumberBlocksTable,
      NumberBlockRow,
      $$NumberBlocksTableFilterComposer,
      $$NumberBlocksTableOrderingComposer,
      $$NumberBlocksTableAnnotationComposer,
      $$NumberBlocksTableCreateCompanionBuilder,
      $$NumberBlocksTableUpdateCompanionBuilder,
      (NumberBlockRow, BaseReferences<_$AppDatabase, $NumberBlocksTable, NumberBlockRow>),
      NumberBlockRow,
      PrefetchHooks Function()
    >;

class $AppDatabaseManager {
  final _$AppDatabase _db;
  $AppDatabaseManager(this._db);
  $$EventSnapshotsTableTableManager get eventSnapshots => $$EventSnapshotsTableTableManager(_db, _db.eventSnapshots);
  $$ProductsTableTableManager get products => $$ProductsTableTableManager(_db, _db.products);
  $$MembersTableTableManager get members => $$MembersTableTableManager(_db, _db.members);
  $$PromotionsTableTableManager get promotions => $$PromotionsTableTableManager(_db, _db.promotions);
  $$PendingOrdersTableTableManager get pendingOrders => $$PendingOrdersTableTableManager(_db, _db.pendingOrders);
  $$ReceiptsTableTableManager get receipts => $$ReceiptsTableTableManager(_db, _db.receipts);
  $$CartDraftsTableTableManager get cartDrafts => $$CartDraftsTableTableManager(_db, _db.cartDrafts);
  $$NumberBlocksTableTableManager get numberBlocks => $$NumberBlocksTableTableManager(_db, _db.numberBlocks);
}
