package com.eventmpos.auth

import org.springframework.beans.factory.annotation.Value
import org.springframework.stereotype.Service
import org.springframework.transaction.annotation.Transactional
import java.security.MessageDigest
import java.security.SecureRandom
import java.time.Instant
import java.time.temporal.ChronoUnit
import java.util.Base64
import java.util.UUID

/**
 * Manages opaque refresh tokens with rotation + revocation.
 * The raw token is returned to the client exactly once; only its hash is persisted.
 */
@Service
class RefreshTokenService(
    private val repository: RefreshTokenRepository,
    @Value("\${app.jwt.refresh-token-ttl-days}") private val refreshTtlDays: Long,
) {
    private val random = SecureRandom()

    fun issue(userId: UUID): String {
        val raw = newRawToken()
        val token = RefreshToken(
            userId = userId,
            tokenHash = hash(raw),
            expiresAt = Instant.now().plus(refreshTtlDays, ChronoUnit.DAYS),
        )
        repository.save(token)
        return raw
    }

    /**
     * Validates a presented refresh token and rotates it: the old token is revoked and a new one
     * issued. Returns the owning userId + the new raw refresh token, or null if invalid/expired.
     */
    @Transactional
    fun rotate(rawToken: String): Rotation? {
        val existing = repository.findByTokenHash(hash(rawToken)) ?: return null
        if (existing.revoked || existing.expiresAt.isBefore(Instant.now())) return null
        existing.revoked = true
        val newRaw = issue(existing.userId)
        return Rotation(existing.userId, newRaw)
    }

    @Transactional
    fun revokeAllForUser(userId: UUID) = repository.revokeAllForUser(userId)

    private fun newRawToken(): String {
        val bytes = ByteArray(32)
        random.nextBytes(bytes)
        return Base64.getUrlEncoder().withoutPadding().encodeToString(bytes)
    }

    private fun hash(raw: String): String {
        val digest = MessageDigest.getInstance("SHA-256").digest(raw.toByteArray())
        return Base64.getEncoder().encodeToString(digest)
    }

    data class Rotation(val userId: UUID, val newRefreshToken: String)
}
