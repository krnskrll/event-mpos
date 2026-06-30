package com.eventmpos.promotion

import jakarta.validation.constraints.NotBlank
import jakarta.validation.constraints.NotNull
import java.math.BigDecimal
import java.time.Instant
import java.util.UUID

data class PromotionResponse(
    val id: UUID,
    val eventId: UUID,
    val name: String,
    val type: PromotionType,
    val percentOff: BigDecimal?,
    val amountOff: BigDecimal?,
    val scope: PromotionScope,
    val productIds: List<UUID>,
    val minQty: Int?,
    val minSubtotal: BigDecimal?,
    val startsAt: Instant?,
    val endsAt: Instant?,
    val active: Boolean,
    val priority: Int,
)

data class PromotionRequest(
    @field:NotBlank val name: String,
    @field:NotNull val type: PromotionType,
    val percentOff: BigDecimal? = null,
    val amountOff: BigDecimal? = null,
    val scope: PromotionScope = PromotionScope.ALL,
    val productIds: List<UUID> = emptyList(),
    val minQty: Int? = null,
    val minSubtotal: BigDecimal? = null,
    val startsAt: Instant? = null,
    val endsAt: Instant? = null,
    val active: Boolean = true,
    val priority: Int = 0,
)

/** Compact promotion shape carried in the event snapshot (client mirrors the engine on these). */
data class PromotionSnapshotDto(
    val id: UUID,
    val name: String,
    val type: PromotionType,
    val percentOff: BigDecimal?,
    val amountOff: BigDecimal?,
    val scope: PromotionScope,
    val productIds: List<UUID>,
    val minQty: Int?,
    val minSubtotal: BigDecimal?,
    val startsAt: Instant?,
    val endsAt: Instant?,
)

// ---- quote (authoritative preview / order-submit application) ----

data class QuoteLine(
    @field:NotNull val productId: UUID,
    val qty: Int,
)

data class QuoteRequest(val lines: List<QuoteLine>)

data class QuoteResponse(
    val subtotal: BigDecimal,
    val discount: BigDecimal,
    val total: BigDecimal,
    val appliedPromotionId: UUID?,
    val appliedPromotionName: String?,
)

fun Promotion.toResponse() = PromotionResponse(
    id, eventId, name, type, percentOff, amountOff, scope, productIds, minQty, minSubtotal,
    startsAt, endsAt, active, priority,
)

fun Promotion.toSnapshotDto() = PromotionSnapshotDto(
    id, name, type, percentOff, amountOff, scope, productIds, minQty, minSubtotal, startsAt, endsAt,
)
