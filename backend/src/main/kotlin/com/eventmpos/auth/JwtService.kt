package com.eventmpos.auth

import com.eventmpos.user.User
import io.jsonwebtoken.Claims
import io.jsonwebtoken.Jwts
import io.jsonwebtoken.security.Keys
import org.springframework.beans.factory.annotation.Value
import org.springframework.stereotype.Service
import java.util.Date
import javax.crypto.SecretKey

/**
 * Issues and verifies short-lived stateless access tokens (JWT).
 * Refresh tokens are handled separately and opaquely by [RefreshTokenService].
 */
@Service
class JwtService(
    @Value("\${app.jwt.secret}") secret: String,
    @Value("\${app.jwt.access-token-ttl-minutes}") private val accessTtlMinutes: Long,
) {
    private val key: SecretKey = Keys.hmacShaKeyFor(io.jsonwebtoken.io.Decoders.BASE64.decode(secret))

    fun generateAccessToken(user: User): String {
        val now = Date()
        val expiry = Date(now.time + accessTtlMinutes * 60_000)
        return Jwts.builder()
            .subject(user.id.toString())
            .claim("email", user.email)
            .claim("role", user.role.name)
            .claim("name", user.fullName)
            .issuedAt(now)
            .expiration(expiry)
            .signWith(key)
            .compact()
    }

    /** Returns the parsed claims, or null if the token is invalid/expired/tampered. */
    fun parse(token: String): Claims? =
        try {
            Jwts.parser().verifyWith(key).build().parseSignedClaims(token).payload
        } catch (ex: Exception) {
            null
        }

    fun accessTokenTtlSeconds(): Long = accessTtlMinutes * 60
}
