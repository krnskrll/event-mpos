package com.eventmpos.stock

import jakarta.validation.constraints.PositiveOrZero
import java.util.UUID

/** Admin allocation view: one row per product with warehouse + event figures. */
data class AllocationRow(
    val productId: UUID,
    val sku: String,
    val name: String,
    val warehouseStock: Int,
    val allocatedQty: Int,
    val soldQty: Int,
    val remaining: Int,
)

data class SetAllocationRequest(
    @field:PositiveOrZero val allocatedQty: Int,
)

/** Warehouse remaining-stock view (near real-time via polling; live push arrives in EPIC-09). */
data class StockLevelRow(
    val productId: UUID,
    val sku: String,
    val name: String,
    val allocatedQty: Int,
    val soldQty: Int,
    val remaining: Int,
    val lowStock: Boolean,
)

/** Compact level for the BA device poll. */
data class BaStockLevel(
    val productId: UUID,
    val remaining: Int,
    val lowStock: Boolean,
)
