package com.eventmpos.admin

import com.eventmpos.audit.AuditAction
import com.eventmpos.audit.AuditService
import com.eventmpos.event.Event
import com.eventmpos.event.EventRepository
import com.eventmpos.event.EventStatus
import com.eventmpos.common.ApiException
import com.eventmpos.order.CustomerInfo
import com.eventmpos.order.OrderLineRequest
import com.eventmpos.order.OrderService
import com.eventmpos.order.OrderStatus
import com.eventmpos.order.SubmitOrderRequest
import com.eventmpos.product.Product
import com.eventmpos.product.ProductRepository
import com.eventmpos.report.ReportService
import com.eventmpos.stock.StockService
import com.eventmpos.sync.SyncConfigService
import com.eventmpos.sync.SyncType
import com.eventmpos.user.CreateUserRequest
import com.eventmpos.user.Role
import com.eventmpos.user.UpdateUserRequest
import com.eventmpos.user.UserService
import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Assertions.assertFalse
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
        "spring.datasource.url=jdbc:h2:mem:admintest;MODE=PostgreSQL;DB_CLOSE_DELAY=-1",
        "spring.datasource.driver-class-name=org.h2.Driver",
        "spring.datasource.username=sa",
        "spring.datasource.password=",
        "spring.jpa.hibernate.ddl-auto=create-drop",
        "app.seed.enabled=false",
    ],
)
class BackofficeAdminTest(
    @Autowired val userService: UserService,
    @Autowired val syncConfigService: SyncConfigService,
    @Autowired val reportService: ReportService,
    @Autowired val auditService: AuditService,
    @Autowired val orderService: OrderService,
    @Autowired val eventRepository: EventRepository,
    @Autowired val productRepository: ProductRepository,
    @Autowired val stockService: StockService,
) {

    @Test
    fun `admin can edit role and deactivate a user`() {
        val created = userService.createUser(
            CreateUserRequest("edit-${UUID.randomUUID().toString().take(6)}@x.io", "password1", "Case Worker", Role.BA),
        )
        val updated = userService.updateUser(created.id, UpdateUserRequest(role = Role.WAREHOUSE, active = false))
        assertEquals(Role.WAREHOUSE, updated.role)
        assertFalse(updated.active)
        assertEquals("Case Worker", updated.fullName) // untouched fields preserved
    }

    @Test
    fun `sync config update validates cron and persists`() {
        val cfg = syncConfigService.update(SyncType.ERP, enabled = true, cron = "0 0 */2 * * *")
        assertTrue(cfg.enabled)
        assertEquals("0 0 */2 * * *", cfg.cron)
        // A bad cron is rejected (never breaks the scheduler).
        assertThrows(ApiException::class.java) {
            syncConfigService.update(SyncType.MEMBER, enabled = true, cron = "not a cron")
        }
    }

    @Test
    fun `end-of-event report totals sales and shows allocated vs remaining`() {
        val event = eventRepository.save(Event(code = "EVRPT", name = "Report Show", status = EventStatus.ACTIVE))
        val p = productRepository.save(Product(sku = "RPT-1", name = "Serum", price = BigDecimal("50.00"), warehouseStock = 500))
        stockService.forceAllocate(event.id, p.id, 10)
        val ba = UUID.randomUUID()
        orderService.submit(event.id, ba, SubmitOrderRequest("rpt-1", CustomerInfo("Aisha"), listOf(OrderLineRequest(p.id, 3))))
        orderService.submit(event.id, ba, SubmitOrderRequest("rpt-2", CustomerInfo("Bryan"), listOf(OrderLineRequest(p.id, 2))))

        val report = reportService.endOfEvent(event.id)
        assertEquals(2, report.totalOrders)
        assertEquals(5, report.totalItems)
        assertEquals(BigDecimal("250.00"), report.totalRevenue)

        val row = report.products.single { it.productId == p.id }
        assertEquals(5, row.unitsSold)
        assertEquals(BigDecimal("250.00"), row.revenue)
        assertEquals(10, row.allocated)
        assertEquals(5, row.sold)
        assertEquals(5, row.remaining)

        val csv = reportService.toCsv(report)
        assertTrue(csv.contains("SKU,Product,Units Sold,Revenue,Allocated,Sold,Remaining"))
        assertTrue(csv.contains("RPT-1,Serum,5,250.00,10,5,5"))
    }

    @Test
    fun `order activity is captured in the append-only audit log`() {
        val event = eventRepository.save(Event(code = "EVAUD", name = "Audit Show", status = EventStatus.ACTIVE))
        val p = productRepository.save(Product(sku = "AUD-1", name = "Cream", price = BigDecimal("20.00"), warehouseStock = 100))
        stockService.forceAllocate(event.id, p.id, 50)
        val ba = UUID.randomUUID()
        val conf = orderService.submit(event.id, ba, SubmitOrderRequest("aud-1", CustomerInfo("Cara"), listOf(OrderLineRequest(p.id, 1))))
        orderService.advanceStatus(conf.id, OrderStatus.PACKING, ba)

        val trail = auditService.forOrder(conf.id)
        assertEquals(2, trail.size)
        assertEquals(AuditAction.ORDER_CREATED, trail[0].action) // FIFO (ascending)
        assertEquals(AuditAction.STATUS_CHANGED, trail[1].action)
        assertEquals("NEW", trail[1].fromStatus)
        assertEquals("PACKING", trail[1].toStatus)

        assertEquals(2, auditService.forEvent(event.id).size)
    }
}
