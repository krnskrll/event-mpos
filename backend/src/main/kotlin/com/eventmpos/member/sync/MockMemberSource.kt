package com.eventmpos.member.sync

import com.eventmpos.member.MemberTier
import org.springframework.boot.autoconfigure.condition.ConditionalOnProperty
import org.springframework.stereotype.Component

/**
 * Default dev member source. QR content is "MBR-xxxxx" — the exact string a customer's QR encodes.
 * Active when `app.member.source` is unset or "mock".
 */
@Component
@ConditionalOnProperty(name = ["app.member.source"], havingValue = "mock", matchIfMissing = true)
class MockMemberSource : MemberSource {

    override val description = "mock"

    override fun fetchMembers(): List<MemberRecord> = listOf(
        MemberRecord("M00001", "MBR-00001", "Aisha Rahman", MemberTier.PLATINUM, "Central", true),
        MemberRecord("M00002", "MBR-00002", "Bryan Tan", MemberTier.GOLD, "Central", true),
        MemberRecord("M00003", "MBR-00003", "Chloe Lim", MemberTier.SILVER, "North", true),
        MemberRecord("M00004", "MBR-00004", "Devan Nair", MemberTier.STANDARD, "South", true),
        MemberRecord("M00005", "MBR-00005", "Elena Wong", MemberTier.GOLD, "Central", true),
        // an inactive member — should be excluded from device cache / lookups
        MemberRecord("M00006", "MBR-00006", "Former Customer", MemberTier.STANDARD, "East", false),
    )
}
