package com.eventmpos.product

import jakarta.validation.Valid
import jakarta.validation.constraints.NotBlank
import jakarta.validation.constraints.NotNull
import jakarta.validation.constraints.PositiveOrZero
import java.math.BigDecimal
import java.time.Instant
import java.util.UUID

data class BarcodeDto(val barcode: String, val label: String?)

/** A variant as seen by admin/web — carries its effective (charged) price. */
data class VariantDto(
    val id: UUID,
    val name: String,
    val sku: String,
    val price: BigDecimal,
    val active: Boolean,
)

data class ProductResponse(
    val id: UUID,
    val sku: String,
    val name: String,
    val brand: String?,
    val price: BigDecimal,
    val imageUrl: String?,
    val source: ProductSource,
    val active: Boolean,
    val warehouseStock: Int,
    val barcodes: List<BarcodeDto>,
    val variants: List<VariantDto>,
    val updatedAt: Instant,
)

data class BarcodeInput(
    @field:NotBlank val barcode: String,
    val label: String? = null,
)

data class CreateProductRequest(
    @field:NotBlank val sku: String,
    @field:NotBlank val name: String,
    val brand: String? = null,
    @field:NotNull @field:PositiveOrZero val price: BigDecimal,
    val imageUrl: String? = null,
    @field:Valid val barcodes: List<BarcodeInput> = emptyList(),
)

/** Partial edit. Null fields are left unchanged; a non-null [barcodes] replaces the whole set. */
data class UpdateProductRequest(
    val name: String? = null,
    val brand: String? = null,
    @field:PositiveOrZero val price: BigDecimal? = null,
    val imageUrl: String? = null,
    val active: Boolean? = null,
    @field:Valid val barcodes: List<BarcodeInput>? = null,
)

/** A catalogue product plus whether it's available for a given event (admin event-product screen). */
data class EventProductResponse(
    val product: ProductResponse,
    val available: Boolean,
)

data class SetAvailabilityRequest(
    @field:NotNull val available: Boolean,
)

/** Compact variant shape for the device snapshot. */
data class VariantSnapshotDto(
    val id: UUID,
    val name: String,
    val price: BigDecimal,
)

/** Compact product shape carried in the event snapshot to the Android device. */
data class ProductSnapshotDto(
    val id: UUID,
    val sku: String,
    val name: String,
    val brand: String?,
    val price: BigDecimal,
    val imageUrl: String?,
    val barcodes: List<String>,
    val variants: List<VariantSnapshotDto> = emptyList(),
    /** Remaining event stock at snapshot time; the device polls /stock for live updates. */
    val remainingStock: Int = 0,
)

fun Product.toResponse() = ProductResponse(
    id = id,
    sku = sku,
    name = name,
    brand = brand,
    price = price,
    imageUrl = imageUrl,
    source = source,
    active = active,
    warehouseStock = warehouseStock,
    barcodes = barcodes.map { BarcodeDto(it.barcode, it.label) },
    variants = variants.map { VariantDto(it.id, it.name, it.sku, it.effectivePrice(), it.active) },
    updatedAt = updatedAt,
)

fun Product.toSnapshotDto() = ProductSnapshotDto(
    id = id,
    sku = sku,
    name = name,
    brand = brand,
    price = price,
    imageUrl = imageUrl,
    barcodes = barcodes.map { it.barcode },
    variants = variants.filter { it.active }.map { VariantSnapshotDto(it.id, it.name, it.effectivePrice()) },
)
