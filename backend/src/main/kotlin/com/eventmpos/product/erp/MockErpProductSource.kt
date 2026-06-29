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

    override fun fetchProducts(): List<ErpProduct> = listOf(
        ErpProduct("SKU-1001", "Hydra Glow Serum 30ml", "LumièreCo", BigDecimal("129.00"),
            "https://picsum.photos/seed/serum/300", listOf("8990001001" to "single", "18990001001" to "box of 6")),
        ErpProduct("SKU-1002", "Velvet Matte Lipstick — Rose", "LumièreCo", BigDecimal("79.00"),
            "https://picsum.photos/seed/lipstick/300", listOf("8990001002" to null)),
        ErpProduct("SKU-1003", "Daily Mineral Sunscreen SPF50", "DermaPure", BigDecimal("99.00"),
            "https://picsum.photos/seed/spf/300", listOf("8990001003" to "single", "28990001003" to "twin pack")),
        ErpProduct("SKU-1004", "Nourishing Hair Oil 100ml", "Botané", BigDecimal("65.00"),
            "https://picsum.photos/seed/hairoil/300", listOf("8990001004" to null)),
        ErpProduct("SKU-1005", "Gentle Foaming Cleanser 150ml", "DermaPure", BigDecimal("55.00"),
            "https://picsum.photos/seed/cleanser/300", listOf("8990001005" to null)),
    )
}
