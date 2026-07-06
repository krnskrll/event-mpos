package com.eventmpos.auth

import jakarta.validation.Valid
import org.springframework.security.core.annotation.AuthenticationPrincipal
import org.springframework.web.bind.annotation.PostMapping
import org.springframework.web.bind.annotation.RequestBody
import org.springframework.web.bind.annotation.RequestMapping
import org.springframework.web.bind.annotation.RestController

@RestController
@RequestMapping("/api/auth")
class AuthController(
    private val authService: AuthService,
) {

    @PostMapping("/login")
    fun login(@Valid @RequestBody request: LoginRequest): TokenResponse =
        authService.login(request.email, request.password)

    @PostMapping("/refresh")
    fun refresh(@Valid @RequestBody request: RefreshRequest): TokenResponse =
        authService.refresh(request.refreshToken)

    /** Revokes all refresh tokens for the caller. The access token expires on its own shortly after. */
    @PostMapping("/logout")
    fun logout(@AuthenticationPrincipal principal: AuthPrincipal) =
        authService.logout(principal.userId)
}
