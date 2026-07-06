package com.eventmpos.mpos.data.auth

import com.eventmpos.mpos.domain.auth.AssignedEvent
import com.eventmpos.mpos.domain.auth.AuthUser
import com.eventmpos.mpos.domain.auth.EventStatus
import com.eventmpos.mpos.domain.auth.Role
import kotlinx.serialization.Serializable

@Serializable
data class LoginRequestDto(val email: String, val password: String)

@Serializable
data class RefreshRequestDto(val refreshToken: String)

@Serializable
data class UserSummaryDto(
    val id: String,
    val email: String,
    val fullName: String,
    val role: String,
)

@Serializable
data class TokenResponseDto(
    val accessToken: String,
    val refreshToken: String,
    val expiresInSeconds: Long,
    val user: UserSummaryDto,
)

@Serializable
data class EventResponseDto(
    val id: String,
    val code: String,
    val name: String,
    val startDate: String? = null,
    val endDate: String? = null,
    val status: String = "DRAFT",
)

/**
 * The snapshot endpoint returns one of two JSON shapes:
 *  - full snapshot: { event, version, products, ... }
 *  - "nothing changed": { changed: false, version }
 * Both decode into this one DTO (unknown keys ignored); [changed] is null for the full shape.
 */
@Serializable
data class ProductDto(
    val id: String,
    val sku: String,
    val name: String,
    val brand: String? = null,
    val price: Double = 0.0,
    val imageUrl: String? = null,
    val barcodes: List<String> = emptyList(),
    val remainingStock: Int = 0,
)

/** Live stock level for the periodic /stock poll. */
@Serializable
data class StockLevelDto(
    val productId: String,
    val remaining: Int = 0,
    val lowStock: Boolean = false,
)

@Serializable
data class MemberDto(
    val qrCode: String,
    val name: String,
    val tier: String = "STANDARD",
)

/** Online QR-lookup response (api/me/members/by-qr). */
@Serializable
data class MemberLookupDto(
    val qrCode: String,
    val name: String,
    val tier: String = "STANDARD",
)

@Serializable
data class PromotionDto(
    val id: String,
    val name: String,
    val type: String,
    val percentOff: Double? = null,
    val amountOff: Double? = null,
    val scope: String = "ALL",
    val productIds: List<String> = emptyList(),
    val minQty: Int? = null,
    val minSubtotal: Double? = null,
    val startsAt: String? = null,
    val endsAt: String? = null,
)

@Serializable
data class SnapshotEnvelopeDto(
    val version: String,
    val changed: Boolean? = null,
    val event: EventResponseDto? = null,
    val products: List<ProductDto> = emptyList(),
    val members: List<MemberDto> = emptyList(),
    val promotions: List<PromotionDto> = emptyList(),
)

@Serializable
data class SnapshotHashDto(
    val snapshotHash: String,
    val checkedAt: String? = null,
)

fun UserSummaryDto.toDomain() = AuthUser(id, email, fullName, Role.valueOf(role))

fun EventResponseDto.toDomain() = AssignedEvent(
    id = id,
    code = code,
    name = name,
    startDate = startDate,
    endDate = endDate,
    status = runCatching { EventStatus.valueOf(status) }.getOrDefault(EventStatus.DRAFT),
)
