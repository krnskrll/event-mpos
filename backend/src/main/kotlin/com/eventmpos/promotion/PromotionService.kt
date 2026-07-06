package com.eventmpos.promotion

import com.eventmpos.common.ApiException
import com.eventmpos.product.ProductRepository
import org.springframework.http.HttpStatus
import org.springframework.stereotype.Service
import org.springframework.transaction.annotation.Transactional
import java.math.BigDecimal
import java.time.Instant
import java.util.UUID

@Service
class PromotionService(
    private val promotionRepository: PromotionRepository,
    private val productRepository: ProductRepository,
) {

    fun listForEvent(eventId: UUID): List<PromotionResponse> =
        promotionRepository.findAllByEventId(eventId).sortedByDescending { it.priority }.map { it.toResponse() }

    fun snapshotPromotions(eventId: UUID): List<PromotionSnapshotDto> =
        promotionRepository.findAllByEventIdAndActiveTrue(eventId).map { it.toSnapshotDto() }

    /** Config fingerprint (no time) folded into the snapshot version so edits re-sync devices. */
    fun fingerprint(eventId: UUID): String =
        promotionRepository.findAllByEventId(eventId).sortedBy { it.id }
            .joinToString(",") { "${it.id}:${it.type}:${it.percentOff}:${it.amountOff}:${it.minQty}:${it.minSubtotal}:${it.active}:${it.startsAt}:${it.endsAt}" }

    @Transactional
    fun create(eventId: UUID, req: PromotionRequest): PromotionResponse {
        validate(req)
        val promo = Promotion(eventId = eventId, name = req.name.trim(), type = req.type)
        apply(promo, req)
        return promotionRepository.save(promo).toResponse()
    }

    @Transactional
    fun update(id: UUID, req: PromotionRequest): PromotionResponse {
        validate(req)
        val promo = promotionRepository.findById(id).orElseThrow {
            ApiException(HttpStatus.NOT_FOUND, "Promotion not found")
        }
        promo.name = req.name.trim()
        promo.type = req.type
        apply(promo, req)
        promo.updatedAt = Instant.now()
        return promotionRepository.save(promo).toResponse()
    }

    @Transactional
    fun delete(id: UUID) {
        if (!promotionRepository.existsById(id)) throw ApiException(HttpStatus.NOT_FOUND, "Promotion not found")
        promotionRepository.deleteById(id)
    }

    /**
     * Authoritative quote for a cart: prices come from the DB, the shared engine picks the single
     * best promotion. EPIC-07's order submit reuses this so the charged discount is server-decided.
     */
    fun quote(eventId: UUID, lines: List<QuoteLine>): QuoteResponse {
        val cart = lines.filter { it.qty > 0 }.mapNotNull { line ->
            productRepository.findById(line.productId).orElse(null)?.let { p ->
                CartLine(p.id, p.price, line.qty)
            }
        }
        val subtotal = cart.fold(BigDecimal.ZERO) { acc, l -> acc + l.unitPrice.multiply(BigDecimal(l.qty)) }
            .setScale(2)
        val promotions = promotionRepository.findAllByEventIdAndActiveTrue(eventId)
        val applied = PromotionEngine.bestPromotion(cart, promotions, Instant.now())
        val discount = applied?.discount ?: BigDecimal.ZERO.setScale(2)
        return QuoteResponse(
            subtotal = subtotal,
            discount = discount,
            total = subtotal.subtract(discount).max(BigDecimal.ZERO).setScale(2),
            appliedPromotionId = applied?.promotion?.id,
            appliedPromotionName = applied?.promotion?.name,
        )
    }

    private fun apply(promo: Promotion, req: PromotionRequest) {
        promo.percentOff = req.percentOff
        promo.amountOff = req.amountOff
        promo.scope = req.scope
        promo.productIds = req.productIds
        promo.minQty = req.minQty
        promo.minSubtotal = req.minSubtotal
        promo.startsAt = req.startsAt
        promo.endsAt = req.endsAt
        promo.active = req.active
        promo.priority = req.priority
    }

    private fun validate(req: PromotionRequest) {
        when (req.type) {
            PromotionType.PERCENTAGE -> if (req.percentOff == null || req.percentOff <= BigDecimal.ZERO)
                throw ApiException(HttpStatus.BAD_REQUEST, "percentOff required for a percentage promotion")
            PromotionType.FIXED -> if (req.amountOff == null || req.amountOff <= BigDecimal.ZERO)
                throw ApiException(HttpStatus.BAD_REQUEST, "amountOff required for a fixed promotion")
            PromotionType.BUNDLE -> {
                if (req.amountOff == null || req.amountOff <= BigDecimal.ZERO)
                    throw ApiException(HttpStatus.BAD_REQUEST, "amountOff required for a bundle promotion")
                if (req.minQty == null || req.minQty < 1)
                    throw ApiException(HttpStatus.BAD_REQUEST, "minQty (bundle size) required for a bundle promotion")
            }
        }
        if (req.scope == PromotionScope.SPECIFIC && req.productIds.isEmpty())
            throw ApiException(HttpStatus.BAD_REQUEST, "SPECIFIC scope needs at least one product")
    }
}
