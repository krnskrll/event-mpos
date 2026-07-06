package com.eventmpos.auth

import com.eventmpos.user.Role
import jakarta.validation.constraints.Email
import jakarta.validation.constraints.NotBlank
import java.util.UUID

data class LoginRequest(
    @field:Email @field:NotBlank val email: String,
    @field:NotBlank val password: String,
)

data class RefreshRequest(
    @field:NotBlank val refreshToken: String,
)

/**
 * Returned by /login and /refresh. accessToken is a JWT; refreshToken is opaque and rotated
 * on every refresh. expiresInSeconds tells the client when to proactively refresh.
 */
data class TokenResponse(
    val accessToken: String,
    val refreshToken: String,
    val expiresInSeconds: Long,
    val user: UserSummary,
)

data class UserSummary(
    val id: UUID,
    val email: String,
    val fullName: String,
    val role: Role,
)
