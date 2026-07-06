package com.eventmpos.product.erp

import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty
import org.springframework.stereotype.Component
import java.math.BigDecimal

/**
 * Default dev ERP source — returns a small fixed catalogue so the whole sync pipeline works out of
 * the box without a real ERP. Active when `app.erp.source` is unset or "mock".
 */
@Component
@ConditionalOnProperty(name = ["app.erp.source"], havingValue = "mock", matchIfMissing = true)
class MockErpProductSource : ErpProductSource {

    override val description = "mock"

    // Prices in Rupiah (whole thousands). imageUrl left null — the apps render a branded placeholder
    // rather than an unrepresentative stock photo.
    override fun fetchProducts(): List<ErpProduct> = listOf(
        ErpProduct("SKU-1001", "Hydra Glow Serum 30ml", "LumièreCo", BigDecimal("129000"),
            null, 120, listOf("8990001001" to "single", "18990001001" to "box of 6")),
        ErpProduct("SKU-1002", "Velvet Matte Lipstick — Rose", "LumièreCo", BigDecimal("79000"),
            null, 300, listOf("8990001002" to null)),
        ErpProduct("SKU-1003", "Daily Mineral Sunscreen SPF50", "DermaPure", BigDecimal("99000"),
            null, 80, listOf("8990001003" to "single", "28990001003" to "twin pack")),
        ErpProduct("SKU-1004", "Nourishing Hair Oil 100ml", "Botané", BigDecimal("65000"),
            null, 45, listOf("8990001004" to null)),
        ErpProduct("SKU-1005", "Gentle Foaming Cleanser 150ml", "DermaPure", BigDecimal("55000"),
            null, 200, listOf("8990001005" to null)),
    )
}
