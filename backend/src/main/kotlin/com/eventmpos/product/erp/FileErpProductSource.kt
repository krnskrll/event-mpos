package com.eventmpos.product.erp

import com.eventmpos.common.ApiException
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty
import org.springframework.core.io.DefaultResourceLoader
import org.springframework.http.HttpStatus
import org.springframework.stereotype.Component
import java.math.BigDecimal

/**
 * File-import ERP source. Reads a CSV export (the most common ERP file format; an Excel/xlsx reader
 * could be dropped in the same way without touching anything else).
 *
 * Expected columns: sku,name,brand,price,imageUrl,stock,barcodes
 * where `barcodes` is "code|label;code|label" (label optional). Configure the location with
 * `app.erp.file-path` (a Spring resource: classpath:..., file:..., etc.).
 */
@Component
@ConditionalOnProperty(name = ["app.erp.source"], havingValue = "file")
class FileErpProductSource(
    @org.springframework.beans.factory.annotation.Value("\${app.erp.file-path:classpath:erp/products.csv}")
    private val filePath: String,
) : ErpProductSource {

    override val description get() = "file:$filePath"

    override fun fetchProducts(): List<ErpProduct> {
        val resource = DefaultResourceLoader().getResource(filePath)
        if (!resource.exists()) {
            throw ApiException(HttpStatus.BAD_REQUEST, "ERP file not found at $filePath")
        }
        val lines = resource.inputStream.bufferedReader().readLines().filter { it.isNotBlank() }
        if (lines.isEmpty()) return emptyList()
        // skip header row
        return lines.drop(1).map { line ->
            val cols = line.split(",").map { it.trim() }
            require(cols.size >= 7) { "Malformed CSV row (need 7 columns): $line" }
            ErpProduct(
                sku = cols[0],
                name = cols[1],
                brand = cols[2].ifBlank { null },
                price = cols[3].toBigDecimalOrNull() ?: BigDecimal.ZERO,
                imageUrl = cols[4].ifBlank { null },
                stockQuantity = cols[5].toIntOrNull() ?: 0,
                barcodes = parseBarcodes(cols[6]),
            )
        }
    }

    private fun parseBarcodes(raw: String): List<Pair<String, String?>> =
        raw.split(";").mapNotNull { entry ->
            val parts = entry.split("|")
            val code = parts.getOrNull(0)?.trim().orEmpty()
            if (code.isBlank()) null else code to parts.getOrNull(1)?.trim()?.ifBlank { null }
        }
}
