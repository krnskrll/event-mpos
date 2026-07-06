package com.eventmpos.order

import com.eventmpos.common.ApiException
import com.eventmpos.event.EventRepository
import com.eventmpos.event.EventStatus
import com.eventmpos.product.ProductRepository
import com.eventmpos.promotion.CartLine
import com.eventmpos.promotion.PromotionEngine
import com.eventmpos.promotion.PromotionRepository
import com.eventmpos.stock.StockService
import com.eventmpos.user.UserRepository
import org.springframework.context.ApplicationEventPublisher
import org.springframework.dao.DataIntegrityViolationException
import org.springframework.http.HttpStatus
import org.springframework.stereotype.Service
import org.springframework.transaction.annotation.Transactional
import java.math.BigDecimal
import java.time.Instant
import java.time.LocalDate
import java.time.ZoneId
import java.time.format.DateTimeFormatter
import java.util.UUID

@Service
class OrderService(
    private val orderRepository: OrderRepository,
    private val sequenceRepository: OrderDailySequenceRepository,
    private val productRepository: ProductRepository,
    private val promotionRepository: PromotionRepository,
    private val eventRepository: EventRepository,
    private val stockService: StockService,
    private val userRepository: UserRepository,
    private val auditService: com.eventmpos.audit.AuditService,
    private val events: ApplicationEventPublisher,
) {
    // Daily order-number rollover follows the event locale (Asia/Jakarta), not UTC.
    private val dateFmt = DateTimeFormatter.ofPattern("yyyyMMdd").withZone(ZoneId.of("Asia/Jakarta"))

    /**
     * Create an order ATOMICALLY: idempotency check → validate → decrement stock (pessimistic locks)
     * → apply best promotion → allocate the daily number → persist. Any failure rolls the whole
     * thing back (no half-decremented stock, no order). A retried submit with the same idempotency
     * key returns the original order instead of creating a second one.
     */
    @Transactional
    fun submit(eventId: UUID, baUserId: UUID, req: SubmitOrderRequest): OrderConfirmation {
        // 1) idempotency — double tap / WorkManager retry returns the same order
        orderRepository.findByIdempotencyKey(req.idempotencyKey)?.let {
            return toConfirmation(it, duplicate = true)
        }

        val event = eventRepository.findById(eventId).orElseThrow {
            ApiException(HttpStatus.NOT_FOUND, "Event not found")
        }
        if (event.status != EventStatus.ACTIVE) {
            throw ApiException(HttpStatus.CONFLICT, "Event ${event.code} is ${event.status} — cannot accept orders")
        }

        // 2) build line items from the catalogue (server-authoritative prices)
        val lines = req.lines.filter { it.qty > 0 }
        if (lines.isEmpty()) throw ApiException(HttpStatus.BAD_REQUEST, "Cart is empty")

        val cartLines = mutableListOf<CartLine>()
        val draftItems = lines.map { line ->
            val p = productRepository.findById(line.productId).orElseThrow {
                ApiException(HttpStatus.BAD_REQUEST, "Unknown product ${line.productId}")
            }
            cartLines += CartLine(p.id, p.price, line.qty)
            DraftItem(p.id, p.sku, p.name, p.price, line.qty, p.price.multiply(BigDecimal(line.qty)))
        }

        // 3) atomically decrement allocated stock (holds row locks until this tx commits)
        lines.forEach { stockService.decrement(eventId, it.productId, it.qty) }

        // 4) pricing — apply the single best promotion (same engine as the quote endpoint)
        val subtotal = draftItems.fold(BigDecimal.ZERO) { acc, i -> acc + i.lineTotal }.setScale(2)
        val applied = PromotionEngine.bestPromotion(
            cartLines, promotionRepository.findAllByEventIdAndActiveTrue(eventId), Instant.now(),
        )
        val discount = applied?.discount ?: BigDecimal.ZERO.setScale(2)
        val total = subtotal.subtract(discount).max(BigDecimal.ZERO).setScale(2)

        // 5) race-safe daily order number
        val dateKey = dateFmt.format(Instant.now())
        val number = nextOrderNumber(eventId, dateKey)
        val referenceId = "ORD-$dateKey-${event.code}-${"%03d".format(number)}"

        // 6) persist (idempotency unique-constraint is the final concurrency backstop)
        val order = Order(
            eventId = eventId, baUserId = baUserId, idempotencyKey = req.idempotencyKey,
            orderNumber = number, referenceId = referenceId,
            customerName = req.customer.name.trim(), customerQr = req.customer.qrCode, customerTier = req.customer.tier,
            subtotal = subtotal, discount = discount, total = total,
            promotionId = applied?.promotion?.id, promotionName = applied?.promotion?.name,
        )
        draftItems.forEach { d ->
            order.items += OrderItem(order = order, productId = d.productId, sku = d.sku, name = d.name,
                unitPrice = d.unitPrice, qty = d.qty, lineTotal = d.lineTotal)
        }
        val saved = try {
            orderRepository.saveAndFlush(order)
        } catch (e: DataIntegrityViolationException) {
            // a concurrent submit with the same idempotency key won the race — return its order
            orderRepository.findByIdempotencyKey(req.idempotencyKey)
                ?.let { return toConfirmation(it, duplicate = true) } ?: throw e
        }
        // Append-only audit: who created this order, when (part of the same atomic transaction).
        auditService.recordOrderCreated(eventId, saved.id, saved.referenceId, baUserId)
        // Broadcast to the warehouse board AFTER this transaction commits.
        events.publishEvent(OrderChangedEvent(eventId, toWarehouseView(saved)))
        return toConfirmation(saved, duplicate = false)
    }

    // ---- warehouse board (EPIC-09) ----

    fun warehouseOrders(eventId: UUID, status: OrderStatus?): List<WarehouseOrderView> =
        orderRepository.findAllByEventIdOrderByCreatedAtDesc(eventId)
            .filter { status == null || it.status == status }
            .map { toWarehouseView(it) }

    /** Advance an order one step (NEW→PACKING→READY→COLLECTED). Broadcasts + audits the change. */
    @Transactional
    fun advanceStatus(orderId: UUID, target: OrderStatus, actorUserId: UUID? = null): WarehouseOrderView {
        val order = orderRepository.findById(orderId).orElseThrow {
            ApiException(HttpStatus.NOT_FOUND, "Order not found")
        }
        if (!order.status.canAdvanceTo(target)) {
            throw ApiException(HttpStatus.CONFLICT, "Cannot move order from ${order.status} to $target")
        }
        val from = order.status
        order.status = target
        val saved = orderRepository.save(order)
        auditService.recordStatusChange(order.eventId, order.id, order.referenceId, from.name, target.name, actorUserId)
        val view = toWarehouseView(saved)
        events.publishEvent(OrderChangedEvent(order.eventId, view))
        return view
    }

    private fun toWarehouseView(o: Order): WarehouseOrderView {
        val baName = userRepository.findById(o.baUserId).map { it.fullName }.orElse("BA")
        return WarehouseOrderView(
            id = o.id, orderNumber = o.orderNumber, orderNumberDisplay = "#${"%03d".format(o.orderNumber)}",
            referenceId = o.referenceId, customerName = o.customerName, baName = baName,
            items = o.items.map { WarehouseItem(it.name, it.sku, it.qty) },
            status = o.status, total = o.total, createdAt = o.createdAt,
        )
    }

    fun history(eventId: UUID, baUserId: UUID): List<OrderSummary> =
        orderRepository.findAllByEventIdAndBaUserIdOrderByCreatedAtDesc(eventId, baUserId).map {
            OrderSummary(it.id, "#${"%03d".format(it.orderNumber)}", it.referenceId, it.status,
                it.customerName, it.total, it.items.sumOf { i -> i.qty }, it.createdAt)
        }

    fun get(orderId: UUID): OrderConfirmation =
        orderRepository.findById(orderId).map { toConfirmation(it, false) }
            .orElseThrow { ApiException(HttpStatus.NOT_FOUND, "Order not found") }

    private fun nextOrderNumber(eventId: UUID, dateKey: String): Int {
        val seq = sequenceRepository.findForUpdate(eventId, dateKey) ?: run {
            runCatching { sequenceRepository.saveAndFlush(OrderDailySequence(eventId = eventId, dateKey = dateKey)) }
            sequenceRepository.findForUpdate(eventId, dateKey)
                ?: throw ApiException(HttpStatus.INTERNAL_SERVER_ERROR, "Could not allocate order number")
        }
        seq.lastNumber += 1
        sequenceRepository.save(seq)
        return seq.lastNumber
    }

    private fun toConfirmation(o: Order, duplicate: Boolean): OrderConfirmation {
        val event = eventRepository.findById(o.eventId).orElse(null)
        return OrderConfirmation(
            id = o.id, orderNumber = o.orderNumber, orderNumberDisplay = "#${"%03d".format(o.orderNumber)}",
            referenceId = o.referenceId, status = o.status,
            eventCode = event?.code ?: "", eventName = event?.name ?: "",
            customerName = o.customerName, customerTier = o.customerTier,
            items = o.items.map { OrderItemView(it.productId, it.sku, it.name, it.unitPrice, it.qty, it.lineTotal) },
            subtotal = o.subtotal, discount = o.discount, total = o.total, promotionName = o.promotionName,
            createdAt = o.createdAt,
            cashierNote = "Please proceed to the cashier to complete payment.",
            duplicate = duplicate,
        )
    }

    private data class DraftItem(
        val productId: UUID, val sku: String, val name: String,
        val unitPrice: BigDecimal, val qty: Int, val lineTotal: BigDecimal,
    )
}
