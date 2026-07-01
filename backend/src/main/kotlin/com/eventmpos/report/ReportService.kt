package com.eventmpos.report

import com.eventmpos.common.ApiException
import com.eventmpos.event.EventRepository
import com.eventmpos.event.EventStatus
import com.eventmpos.order.OrderRepository
import com.eventmpos.product.ProductRepository
import com.eventmpos.stock.EventStockAllocationRepository
import org.springframework.http.HttpStatus
import org.springframework.stereotype.Service
import java.math.BigDecimal
import java.time.Instant
import java.util.UUID

data class ProductReportRow(
    val productId: UUID,
    val sku: String,
    val name: String,
    val unitsSold: Int,
    val revenue: BigDecimal,
    val allocated: Int,
    val sold: Int,
    val remaining: Int,
)

data class EndOfEventReport(
    val eventId: UUID,
    val eventCode: String,
    val eventName: String,
    val status: EventStatus,
    val generatedAt: Instant,
    val totalOrders: Int,
    val totalItems: Int,
    val totalRevenue: BigDecimal,
    val totalDiscount: BigDecimal,
    val products: List<ProductReportRow>,
)

/**
 * End-of-event reporting. Aggregates the event's orders (revenue, items sold per product) and joins
 * against the stock allocations (allocated vs remaining). Backed by indexed lookups: orders by
 * `event_id` (idx_order_event), their items by `order_id` (idx_order_item_order), allocations by
 * `event_id` (idx_event_stock_event) — so the query cost scales with the event's own data, not the
 * whole table. Designed to be run after an event CLOSES, but works at any status (status is included).
 */
@Service
class ReportService(
    private val orderRepository: OrderRepository,
    private val allocationRepository: EventStockAllocationRepository,
    private val productRepository: ProductRepository,
    private val eventRepository: EventRepository,
) {
    fun endOfEvent(eventId: UUID): EndOfEventReport {
        val event = eventRepository.findById(eventId).orElseThrow {
            ApiException(HttpStatus.NOT_FOUND, "Event not found")
        }
        val orders = orderRepository.findAllByEventIdOrderByCreatedAtDesc(eventId)

        val totalRevenue = orders.sumOf { it.total }
        val totalDiscount = orders.sumOf { it.discount }
        val totalItems = orders.sumOf { o -> o.items.sumOf { it.qty } }

        // Aggregate units sold + revenue per product from the order lines.
        data class Agg(var sku: String, var name: String, var qty: Int = 0, var revenue: BigDecimal = BigDecimal.ZERO)
        val sales = LinkedHashMap<UUID, Agg>()
        orders.forEach { o ->
            o.items.forEach { i ->
                val agg = sales.getOrPut(i.productId) { Agg(i.sku, i.name) }
                agg.qty += i.qty
                agg.revenue = agg.revenue.add(i.lineTotal)
            }
        }

        val allocations = allocationRepository.findAllByEventId(eventId).associateBy { it.productId }

        // Union: every product that either sold or had stock allocated to the event.
        val productIds = LinkedHashSet<UUID>().apply { addAll(sales.keys); addAll(allocations.keys) }
        val rows = productIds.map { pid ->
            val s = sales[pid]
            val a = allocations[pid]
            val fallback = if (s == null) productRepository.findById(pid).orElse(null) else null
            ProductReportRow(
                productId = pid,
                sku = s?.sku ?: fallback?.sku ?: "—",
                name = s?.name ?: fallback?.name ?: "(unknown)",
                unitsSold = s?.qty ?: 0,
                revenue = (s?.revenue ?: BigDecimal.ZERO).setScale(2),
                allocated = a?.allocatedQty ?: 0,
                sold = a?.soldQty ?: 0,
                remaining = a?.remaining ?: 0,
            )
        }.sortedByDescending { it.unitsSold }

        return EndOfEventReport(
            eventId = event.id,
            eventCode = event.code,
            eventName = event.name,
            status = event.status,
            generatedAt = Instant.now(),
            totalOrders = orders.size,
            totalItems = totalItems,
            totalRevenue = totalRevenue.setScale(2),
            totalDiscount = totalDiscount.setScale(2),
            products = rows,
        )
    }

    /** Render the report as CSV: a small summary block, then one row per product. */
    fun toCsv(report: EndOfEventReport): String {
        fun esc(s: String) = if (s.contains(',') || s.contains('"')) "\"" + s.replace("\"", "\"\"") + "\"" else s
        val sb = StringBuilder()
        sb.appendLine("End-of-Event Report,${esc("${report.eventCode} — ${report.eventName}")}")
        sb.appendLine("Status,${report.status}")
        sb.appendLine("Generated,${report.generatedAt}")
        sb.appendLine("Total Orders,${report.totalOrders}")
        sb.appendLine("Total Items,${report.totalItems}")
        sb.appendLine("Total Revenue,${report.totalRevenue}")
        sb.appendLine("Total Discount,${report.totalDiscount}")
        sb.appendLine()
        sb.appendLine("SKU,Product,Units Sold,Revenue,Allocated,Sold,Remaining")
        report.products.forEach { p ->
            sb.appendLine("${esc(p.sku)},${esc(p.name)},${p.unitsSold},${p.revenue},${p.allocated},${p.sold},${p.remaining}")
        }
        return sb.toString()
    }
}
