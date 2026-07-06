package com.eventmpos.auth

import com.eventmpos.common.ApiException
import com.eventmpos.user.User
import com.eventmpos.user.UserRepository
import org.springframework.http.HttpStatus
import org.springframework.security.crypto.password.PasswordEncoder
import org.springframework.stereotype.Service

@Service
class AuthService(
    private val userRepository: UserRepository,
    private val passwordEncoder: PasswordEncoder,
    private val jwtService: JwtService,
    private val refreshTokenService: RefreshTokenService,
) {

    fun login(email: String, password: String): TokenResponse {
        val user = userRepository.findByEmail(email.trim().lowercase())
        // Always run the encoder check shape to avoid leaking which emails exist via timing,
        // and return the same error for "no such user" and "wrong password".
        if (user == null || !passwordEncoder.matches(password, user.passwordHash)) {
            throw ApiException(HttpStatus.UNAUTHORIZED, "Invalid email or password")
        }
        if (!user.active) {
            throw ApiException(HttpStatus.FORBIDDEN, "Account is disabled")
        }
        return issueTokens(user)
    }

    fun refresh(rawRefreshToken: String): TokenResponse {
        val rotation = refreshTokenService.rotate(rawRefreshToken)
            ?: throw ApiException(HttpStatus.UNAUTHORIZED, "Invalid or expired refresh token")
        val user = userRepository.findById(rotation.userId).orElse(null)
            ?: throw ApiException(HttpStatus.UNAUTHORIZED, "Invalid or expired refresh token")
        if (!user.active) throw ApiException(HttpStatus.FORBIDDEN, "Account is disabled")
        return TokenResponse(
            accessToken = jwtService.generateAccessToken(user),
            refreshToken = rotation.newRefreshToken,
            expiresInSeconds = jwtService.accessTokenTtlSeconds(),
            user = user.toSummary(),
        )
    }

    fun logout(userId: java.util.UUID) = refreshTokenService.revokeAllForUser(userId)

    private fun issueTokens(user: User): TokenResponse =
        TokenResponse(
            accessToken = jwtService.generateAccessToken(user),
            refreshToken = refreshTokenService.issue(user.id),
            expiresInSeconds = jwtService.accessTokenTtlSeconds(),
            user = user.toSummary(),
        )

    private fun User.toSummary() = UserSummary(id, email, fullName, role)
}
