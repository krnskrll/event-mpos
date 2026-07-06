package com.eventmpos.member.sync

import com.eventmpos.member.MemberTier

/** Normalized member record from the external system, regardless of transport (REST or file). */
data class MemberRecord(
    val memberCode: String,
    val qrCode: String,
    val name: String,
    val tier: MemberTier,
    val region: String?,
    val active: Boolean,
    val phone: String? = null,
)

/**
 * Pluggable boundary to the external member system — same swap-without-rewrite design as the ERP
 * product source. Selected by `app.member.source`: mock (default) | file | rest.
 */
interface MemberSource {
    val description: String
    fun fetchMembers(): List<MemberRecord>
}
