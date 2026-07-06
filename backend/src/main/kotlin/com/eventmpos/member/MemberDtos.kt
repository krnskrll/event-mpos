package com.eventmpos.member

import java.time.Instant

/** Full member view (admin directory). */
data class MemberResponse(
    val id: String,
    val memberCode: String,
    val qrCode: String,
    val name: String,
    val tier: MemberTier,
    val region: String?,
    val active: Boolean,
    val updatedAt: Instant,
)

/** Compact member shape for the device cache / snapshot — display + linking only. */
data class MemberSnapshotDto(
    val qrCode: String,
    val name: String,
    val tier: MemberTier,
)

fun Member.toResponse() = MemberResponse(
    id.toString(), memberCode, qrCode, name, tier, region, active, updatedAt,
)

fun Member.toSnapshotDto() = MemberSnapshotDto(qrCode, name, tier)
