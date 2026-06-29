package com.eventmpos.auth

import com.eventmpos.user.Role
import java.util.UUID

/**
 * The authenticated user attached to the SecurityContext on each request.
 * Controllers can inject this via @AuthenticationPrincipal.
 */
data class AuthPrincipal(
    val userId: UUID,
    val email: String,
    val fullName: String,
    val role: Role,
)
