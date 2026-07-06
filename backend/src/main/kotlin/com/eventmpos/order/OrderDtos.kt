package com.eventmpos.order

import jakarta.validation.Valid
import jakarta.validation.constraints.NotBlank
import jakarta.validation.constraints.NotEmpty
import jakarta.validation.constraints.NotNull
import java.math.BigDecimal
import java.time.Instant
import java.util.UUID

data class CustomerInfo(
    @field:NotBlank val name: String,
    val qrCode: String? = null,
    val tier: String? = null,
)

data class OrderLineRequest(
    @field:NotNull val productId: UUID,
    val qty: Int,
    /** Optional chosen variant (shade/size). Null = the base product. */
    val variantId: UUID? = null,
)

data class SubmitOrderRequest(
    /** Client-generated; identical across retries so double-taps don't create duplicates. */
    @field:NotBlank val idempotencyKey: String,
    @field:Valid @field:NotNull val customer: CustomerInfo,
    @field:NotEmpty val lines: List<OrderLineRequest>,
)

data class OrderItemView(
    val productId: UUID,
    val sku: String,
    val name: String,
    val unitPrice: BigDecimal,
    val qty: Int,
    val lineTotal: BigDecimal,
    val variantName: String? = null,
)

/**
 * Confirmation payload — also the source for the 58mm receipt in EPIC-08 (brand/event/order number,
 * customer, line items, discount, total, cashier footer all present).
 */
data class OrderConfirmation(
    val id: UUID,
    val orderNumber: Int,
    val orderNumberDisplay: String,   // "#042"
    val referenceId: String,          // ORD-20260629-EV01-042
    val status: OrderStatus,
    val eventCode: String,
    val eventName: String,
    val customerName: String,
    val customerTier: String?,
    val items: List<OrderItemView>,
    val subtotal: BigDecimal,
    val discount: BigDecimal,
    val total: BigDecimal,
    val promotionName: String?,
    val createdAt: Instant,
    val cashierNote: String,
    val duplicate: Boolean = false,   // true if this submit hit the idempotency guard
)

/** Compact row for order history. */
data class OrderSummary(
    val id: UUID,
    val orderNumberDisplay: String,
    val referenceId: String,
    val status: OrderStatus,
    val customerName: String,
    val total: BigDecimal,
    val itemCount: Int,
    val createdAt: Instant,
)
