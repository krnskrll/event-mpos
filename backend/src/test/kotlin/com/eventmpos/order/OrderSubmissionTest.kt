package com.eventmpos.order

import com.eventmpos.common.ApiException
import com.eventmpos.event.Event
import com.eventmpos.event.EventRepository
import com.eventmpos.event.EventStatus
import com.eventmpos.product.Product
import com.eventmpos.product.ProductRepository
import com.eventmpos.promotion.Promotion
import com.eventmpos.promotion.PromotionRepository
import com.eventmpos.promotion.PromotionType
import com.eventmpos.stock.StockService
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Assertions.assertThrows
import org.junit.jupiter.api.Assertions.assertTrue
import org.junit.jupiter.api.Test
import org.springframework.beans.factory.annotation.Autowired
import org.springframework.boot.test.context.SpringBootTest
import org.springframework.test.context.TestPropertySource
import java.math.BigDecimal
import java.util.UUID

@SpringBootTest
@TestPropertySource(
    properties = [
        "spring.datasource.url=jdbc:h2:mem:ordertest;MODE=PostgreSQL;DB_CLOSE_DELAY=-1",
        "spring.datasource.driver-class-name=org.h2.Driver",
        "spring.datasource.username=sa",
        "spring.datasource.password=",
        "spring.jpa.hibernate.ddl-auto=create-drop",
        "app.seed.enabled=false",
    ],
)
class OrderSubmissionTest(
    @Autowired val orderService: OrderService,
    @Autowired val eventRepository: EventRepository,
    @Autowired val productRepository: ProductRepository,
    @Autowired val promotionRepository: PromotionRepository,
    @Autowired val stockService: StockService,
) {
    private fun setup(allocated: Int = 100): Triple<UUID, Product, UUID> {
        val code = "EV" + UUID.randomUUID().toString().take(4).uppercase()
        val event = eventRepository.save(Event(code = code, name = "Orders", status = EventStatus.ACTIVE))
        val product = productRepository.save(Product(sku = "S-${UUID.randomUUID().toString().take(6)}", name = "Widget", price = BigDecimal("100.00"), warehouseStock = 1000))
        stockService.forceAllocate(event.id, product.id, allocated)
        return Triple(event.id, product, UUID.randomUUID())
    }

    @Test
    fun `submit creates order with number, reference, decremented stock`() {
        val (eventId, product, ba) = setup(allocated = 10)
        val req = SubmitOrderRequest("idem-1", CustomerInfo("Aisha", "MBR-1", "GOLD"),
            listOf(OrderLineRequest(product.id, 3)))
        val conf = orderService.submit(eventId, ba, req)

        assertEquals(1, conf.orderNumber)
        assertTrue(conf.referenceId.startsWith("ORD-") && conf.referenceId.endsWith("-001")) { conf.referenceId }
        assertEquals(BigDecimal("300.00"), conf.subtotal)
        assertEquals(7, stockService.remaining(eventId, product.id)) // 10 - 3
        assertEquals("Aisha", conf.customerName)
    }

    @Test
    fun `idempotent submit returns same order, does not double-decrement`() {
        val (eventId, product, ba) = setup(allocated = 10)
        val req = SubmitOrderRequest("idem-dup", CustomerInfo("Bryan"), listOf(OrderLineRequest(product.id, 2)))
        val first = orderService.submit(eventId, ba, req)
        val second = orderService.submit(eventId, ba, req) // double tap / retry

        assertEquals(first.id, second.id)
        assertTrue(second.duplicate)
        assertEquals(8, stockService.remaining(eventId, product.id)) // decremented once, not twice
    }

    @Test
    fun `daily number increments per order`() {
        val (eventId, product, ba) = setup(allocated = 100)
        val n1 = orderService.submit(eventId, ba, SubmitOrderRequest("k1", CustomerInfo("A"), listOf(OrderLineRequest(product.id, 1)))).orderNumber
        val n2 = orderService.submit(eventId, ba, SubmitOrderRequest("k2", CustomerInfo("B"), listOf(OrderLineRequest(product.id, 1)))).orderNumber
        assertEquals(1, n1)
        assertEquals(2, n2)
    }

    @Test
    fun `oversell is rejected and rolls back`() {
        val (eventId, product, ba) = setup(allocated = 2)
        assertThrows(ApiException::class.java) {
            orderService.submit(eventId, ba, SubmitOrderRequest("k-over", CustomerInfo("C"), listOf(OrderLineRequest(product.id, 5))))
        }
        assertEquals(2, stockService.remaining(eventId, product.id)) // unchanged — rolled back
    }

    @Test
    fun `best promotion is applied at submit`() {
        val (eventId, product, ba) = setup(allocated = 100)
        promotionRepository.save(Promotion(eventId = eventId, name = "10% off", type = PromotionType.PERCENTAGE, percentOff = BigDecimal("10")))
        // 2 x 100 = 200 subtotal, 10% -> 20 discount, total 180
        val conf = orderService.submit(eventId, ba, SubmitOrderRequest("k-promo", CustomerInfo("D"), listOf(OrderLineRequest(product.id, 2))))
        assertEquals(BigDecimal("200.00"), conf.subtotal)
        assertEquals(BigDecimal("20.00"), conf.discount)
        assertEquals(BigDecimal("180.00"), conf.total)
        assertEquals("10% off", conf.promotionName)
    }
}
