package com.eventmpos.mpos.domain.promotion

import kotlinx.serialization.json.Json
import kotlinx.serialization.json.JsonElement
import kotlinx.serialization.json.JsonNull
import kotlinx.serialization.json.JsonObject
import kotlinx.serialization.json.contentOrNull
import kotlinx.serialization.json.int
import kotlinx.serialization.json.intOrNull
import kotlinx.serialization.json.jsonArray
import kotlinx.serialization.json.jsonObject
import kotlinx.serialization.json.jsonPrimitive
import org.junit.Assert.assertTrue
import org.junit.Test
import java.io.File
import java.time.Instant
import kotlin.math.abs

/**
 * Runs the shared fixture at testdata/promotion-cases.json (repo root) against the client engine.
 * backend-go/internal/promotion/parity_test.go runs the same file against the server engine, so a
 * change to only one engine fails CI. Gradle passes the fixture path as the `promotionCases` property.
 */
class PromotionParityTest {

    private fun fixture(): JsonObject {
        val path = System.getProperty("promotionCases")
            ?: error("promotionCases system property not set (see app/build.gradle.kts)")
        return Json.parseToJsonElement(File(path).readText()).jsonObject
    }

    private fun JsonObject.str(key: String): String? = (this[key] as? JsonElement)
        ?.takeUnless { it is JsonNull }?.jsonPrimitive?.contentOrNull

    private fun JsonObject.toPromotion(index: Int) = Promotion(
        id = "p$index",
        name = str("name")!!,
        type = PromotionType.valueOf(str("type")!!),
        percentOff = str("percentOff")?.toDouble(),
        amountOff = str("amountOff")?.toDouble(),
        scope = str("scope")?.let { PromotionScope.valueOf(it) } ?: PromotionScope.ALL,
        productIds = this["productIds"]?.jsonArray?.map { it.jsonPrimitive.content } ?: emptyList(),
        minQty = this["minQty"]?.jsonPrimitive?.intOrNull,
        minSubtotal = str("minSubtotal")?.toDouble(),
        startsAtEpochMs = str("startsAt")?.let { Instant.parse(it).toEpochMilli() },
        endsAtEpochMs = str("endsAt")?.let { Instant.parse(it).toEpochMilli() },
        priority = this["priority"]?.jsonPrimitive?.intOrNull ?: 0,
    )

    @Test
    fun `client engine matches the shared fixture`() {
        val fx = fixture()
        val now = Instant.parse(fx.str("now")!!).toEpochMilli()
        val carts = fx["carts"]!!.jsonObject.mapValues { (_, lines) ->
            lines.jsonArray.map {
                val l = it.jsonObject
                CartLine(l.str("productId")!!, l.str("unitPrice")!!.toDouble(), l["qty"]!!.jsonPrimitive.int)
            }
        }
        val cases = fx["cases"]!!.jsonArray
        assertTrue("fixture has no cases", cases.isNotEmpty())

        val failures = cases.mapNotNull { element ->
            val c = element.jsonObject
            val name = c.str("name")!!
            val cart = carts[c.str("cart")] ?: return@mapNotNull "$name: unknown cart ${c.str("cart")}"
            val promos = c["promotions"]!!.jsonArray.mapIndexed { i, p -> p.jsonObject.toPromotion(i) }
            val expected = c["expected"]!!.jsonObject
            val wantName = expected.str("promotion")
            val wantDiscount = expected.str("discount")!!.toDouble()

            val best = PromotionEngine.bestPromotion(cart, promos, now)
            val ok = if (wantName == null) best == null
            else best != null && best.promotion.name == wantName && abs(best.discount - wantDiscount) < 0.0001
            if (ok) null
            else "$name: chose ${best?.promotion?.name} (${best?.discount}), want $wantName ($wantDiscount)"
        }
        assertTrue("Promotion parity failures:\n" + failures.joinToString("\n"), failures.isEmpty())
    }
}
