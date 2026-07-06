package com.eventmpos.stock

import com.eventmpos.common.ApiException
import com.eventmpos.event.Event
import com.eventmpos.event.EventRepository
import com.eventmpos.event.EventStatus
import com.eventmpos.product.Product
import com.eventmpos.product.ProductRepository
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Test
import org.springframework.beans.factory.annotation.Autowired
import org.springframework.boot.test.context.SpringBootTest
import org.springframework.test.context.TestPropertySource
import java.util.UUID
import java.util.concurrent.ConcurrentLinkedQueue
import java.util.concurrent.CountDownLatch
import java.util.concurrent.Executors
import java.util.concurrent.atomic.AtomicInteger

/**
 * Proves the core EPIC-05 guarantee: even with many BAs submitting at the exact same moment,
 * the allocated stock can never be oversold. Uses a real Postgres-mode (H2) DB and fires many
 * concurrent decrements against a single allocation of 100 units.
 *
 * NOTE: run against Postgres for true SELECT FOR UPDATE semantics; H2 in PostgreSQL mode honors the
 * pessimistic lock + @Version here well enough to assert the invariant. The production path is
 * identical (PESSIMISTIC_WRITE on event_stock_allocations).
 */
@SpringBootTest
@TestPropertySource(
    properties = [
        "spring.datasource.url=jdbc:h2:mem:stocktest;MODE=PostgreSQL;DB_CLOSE_DELAY=-1",
        "spring.datasource.driver-class-name=org.h2.Driver",
        "spring.datasource.username=sa",
        "spring.datasource.password=",
        "spring.jpa.hibernate.ddl-auto=create-drop",
        "app.seed.enabled=false",
    ],
)
class StockConcurrencyTest(
    @Autowired val stockService: StockService,
    @Autowired val allocationRepository: EventStockAllocationRepository,
    @Autowired val productRepository: ProductRepository,
    @Autowired val eventRepository: EventRepository,
) {

    @Test
    fun `concurrent decrements never oversell`() {
        val allocated = 100
        val attempts = 300            // far more than available
        val event = eventRepository.save(Event(code = "EVT-${UUID.randomUUID().toString().take(4)}", name = "Concurrency", status = EventStatus.ACTIVE))
        val product = productRepository.save(Product(sku = "SKU-${UUID.randomUUID().toString().take(6)}", name = "Race Item", warehouseStock = 1000))
        stockService.forceAllocate(event.id, product.id, allocated)

        val pool = Executors.newFixedThreadPool(32)
        val start = CountDownLatch(1)
        val ok = AtomicInteger(0)
        val failed = AtomicInteger(0)
        val errors = ConcurrentLinkedQueue<Throwable>()
        val done = CountDownLatch(attempts)

        repeat(attempts) {
            pool.submit {
                start.await()
                try {
                    stockService.decrement(event.id, product.id, 1)
                    ok.incrementAndGet()
                } catch (e: ApiException) {
                    failed.incrementAndGet() // expected once sold out (409)
                } catch (e: Exception) {
                    errors.add(e)
                } finally {
                    done.countDown()
                }
            }
        }
        start.countDown()             // release all threads at once
        done.await()
        pool.shutdown()

        val finalSold = allocationRepository.findByEventIdAndProductId(event.id, product.id)!!.soldQty

        // Exactly `allocated` succeed, the rest are cleanly rejected, soldQty never exceeds allocation.
        assertEquals(allocated, ok.get(), "exactly $allocated decrements should succeed")
        assertEquals(attempts - allocated, failed.get(), "the rest should be rejected")
        assertEquals(allocated, finalSold, "soldQty must equal allocation, never more (no overselling)")
        assert(errors.isEmpty()) { "unexpected errors: ${errors.map { it.message }}" }
    }
}
