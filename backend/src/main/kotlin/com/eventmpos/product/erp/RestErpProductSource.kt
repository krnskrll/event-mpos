package com.eventmpos.product.erp

import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty
import org.springframework.stereotype.Component
import org.springframework.web.client.RestClient
import java.math.BigDecimal

/**
 * REST-pull ERP source. Calls a configured ERP endpoint that returns product JSON and maps it into
 * [ErpProduct]. Active when `app.erp.source=rest`; configure the URL with `app.erp.rest-url`.
 *
 * The mapping below assumes a simple ERP JSON shape; adjust [ErpProductJson] to match the real ERP.
 */
@Component
@ConditionalOnProperty(name = ["app.erp.source"], havingValue = "rest")
class RestErpProductSource(
    @org.springframework.beans.factory.annotation.Value("\${app.erp.rest-url:}")
    private val restUrl: String,
) : ErpProductSource {

    private val client = RestClient.create()

    override val description get() = "rest:$restUrl"

    override fun fetchProducts(): List<ErpProduct> {
        val body = client.get().uri(restUrl).retrieve()
            .body(Array<ErpProductJson>::class.java) ?: emptyArray()
        return body.map {
            ErpProduct(
                sku = it.sku,
                name = it.name,
                brand = it.brand,
                price = it.price ?: BigDecimal.ZERO,
                imageUrl = it.imageUrl,
                stockQuantity = it.stockQuantity ?: 0,
                barcodes = it.barcodes.map { b -> b.code to b.label },
            )
        }
    }

    data class ErpProductJson(
        val sku: String = "",
        val name: String = "",
        val brand: String? = null,
        val price: BigDecimal? = null,
        val imageUrl: String? = null,
        val stockQuantity: Int? = null,
        val barcodes: List<BarcodeJson> = emptyList(),
    )

    data class BarcodeJson(val code: String = "", val label: String? = null)
}
