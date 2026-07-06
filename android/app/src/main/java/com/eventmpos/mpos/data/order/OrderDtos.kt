package com.eventmpos.mpos.data.order

import kotlinx.serialization.Serializable

@Serializable
data class CustomerInfoDto(val name: String, val qrCode: String? = null, val tier: String? = null)

@Serializable
data class OrderLineDto(val productId: String, val qty: Int)

@Serializable
data class SubmitOrderRequestDto(
    val idempotencyKey: String,
    val customer: CustomerInfoDto,
    val lines: List<OrderLineDto>,
)

@Serializable
data class OrderItemViewDto(
    val productId: String,
    val sku: String,
    val name: String,
    val unitPrice: Double,
    val qty: Int,
    val lineTotal: Double,
)

@Serializable
data class OrderConfirmationDto(
    val id: String,
    val orderNumber: Int,
    val orderNumberDisplay: String,
    val referenceId: String,
    val status: String,
    val eventCode: String = "",
    val eventName: String = "",
    val customerName: String,
    val customerTier: String? = null,
    val items: List<OrderItemViewDto> = emptyList(),
    val subtotal: Double = 0.0,
    val discount: Double = 0.0,
    val total: Double = 0.0,
    val promotionName: String? = null,
    val createdAt: String = "",
    val cashierNote: String = "",
    val duplicate: Boolean = false,
)

@Serializable
data class OrderSummaryDto(
    val id: String,
    val orderNumberDisplay: String,
    val referenceId: String,
    val status: String,
    val customerName: String,
    val total: Double,
    val itemCount: Int,
    val createdAt: String,
)
