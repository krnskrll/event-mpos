package com.eventmpos.auth

import jakarta.persistence.Column
import jakarta.persistence.Entity
import jakarta.persistence.Id
import jakarta.persistence.Index
import jakarta.persistence.Table
import java.time.Instant
import java.util.UUID

/**
 * Persisted refresh token. We store only a SHA-256 hash of the opaque token value, never the raw
 * token — so a DB leak cannot be replayed. Tokens are rotated on every use and can be revoked.
 */
@Entity
@Table(
    name = "refresh_tokens",
    indexes = [Index(name = "idx_refresh_token_hash", columnList = "token_hash", unique = true)],
)
class RefreshToken(
    @Id
    val id: UUID = UUID.randomUUID(),

    @Column(name = "user_id", nullable = false)
    val userId: UUID,

    @Column(name = "token_hash", nullable = false, unique = true)
    val tokenHash: String,

    @Column(name = "expires_at", nullable = false)
    val expiresAt: Instant,

    @Column(nullable = false)
    var revoked: Boolean = false,

    @Column(name = "created_at", nullable = false)
    val createdAt: Instant = Instant.now(),
)
