package com.eventmpos.member

import jakarta.persistence.Column
import jakarta.persistence.Entity
import jakarta.persistence.EnumType
import jakarta.persistence.Enumerated
import jakarta.persistence.Id
import jakarta.persistence.Index
import jakarta.persistence.Table
import java.time.Instant
import java.util.UUID

enum class MemberTier { STANDARD, SILVER, GOLD, PLATINUM }

/**
 * A customer synced from the external member system. Deliberately minimal — only what's needed to
 * identify a customer and link an order (per .cursorrules / EPIC-04: "no financial data").
 *
 * [qrCode] is the exact string encoded in the customer's QR code and is uniquely indexed for fast
 * lookup (PostgreSQL here; Room mirrors the index on the device).
 */
@Entity
@Table(
    name = "members",
    indexes = [Index(name = "idx_member_qr", columnList = "qr_code", unique = true)],
)
class Member(
    @Id
    val id: UUID = UUID.randomUUID(),

    /** Stable external member id (sync upsert key). */
    @Column(name = "member_code", nullable = false, unique = true)
    var memberCode: String,

    @Column(name = "qr_code", nullable = false, unique = true)
    var qrCode: String,

    @Column(nullable = false)
    var name: String,

    @Enumerated(EnumType.STRING)
    @Column(nullable = false)
    var tier: MemberTier = MemberTier.STANDARD,

    /** Optional region — lets sync filter large datasets (e.g. only this market's members). */
    var region: String? = null,

    @Column(nullable = false)
    var active: Boolean = true,

    @Column(name = "updated_at", nullable = false)
    var updatedAt: Instant = Instant.now(),
)
