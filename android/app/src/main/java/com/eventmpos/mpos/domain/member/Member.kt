package com.eventmpos.mpos.domain.member

import kotlinx.serialization.Serializable

/** Minimal customer for identification + order linking. */
data class Member(
    val qrCode: String,
    val name: String,
    val tier: String,
)

/** How a customer was identified — surfaced to the BA so they know if it was a confirmed match. */
enum class CustomerSource { LOCAL_CACHE, ONLINE_LOOKUP, MANUAL }

/** The customer attached to the current selling session. */
@Serializable
data class SelectedCustomer(
    val name: String,
    val tier: String?,
    val qrCode: String?,
    val source: CustomerSource,
)

interface MemberRepository {
    /**
     * Identify a scanned QR: local Room cache first (instant, offline), then an online fallback if
     * connected. Returns null when the QR is unknown everywhere (BA then types a name manually).
     */
    suspend fun identify(qrCode: String): Member?

    suspend fun cachedCount(): Int
}
