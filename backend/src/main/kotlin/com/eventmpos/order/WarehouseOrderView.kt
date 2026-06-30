package com.eventmpos.order

import jakarta.validation.constraints.NotNull
import java.math.BigDecimal
import java.time.Instant
import java.util.UUID

/** Order as the warehouse board sees it — big, glanceable, with everything needed to pack. */
data class WarehouseOrderView(
    val id: UUID,
    val orderNumber: Int,
    val orderNumberDisplay: String,
    val referenceId: String,
    val customerName: String,
    val baName: String,
    val items: List<WarehouseItem>,
    val status: OrderStatus,
    val total: BigDecimal,
    val createdAt: Instant,
)

data class WarehouseItem(val name: String, val sku: String, val qty: Int)

data class AdvanceStatusRequest(
    @field:NotNull val status: OrderStatus,
)
