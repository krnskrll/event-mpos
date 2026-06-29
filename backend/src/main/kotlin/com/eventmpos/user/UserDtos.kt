package com.eventmpos.user

import jakarta.validation.constraints.Email
import jakarta.validation.constraints.NotBlank
import jakarta.validation.constraints.NotNull
import jakarta.validation.constraints.Size
import java.time.Instant
import java.util.UUID

data class CreateUserRequest(
    @field:Email @field:NotBlank val email: String,
    @field:NotBlank @field:Size(min = 8, message = "Password must be at least 8 characters") val password: String,
    @field:NotBlank val fullName: String,
    @field:NotNull val role: Role,
)

data class UserResponse(
    val id: UUID,
    val email: String,
    val fullName: String,
    val role: Role,
    val active: Boolean,
    val createdAt: Instant,
)

fun User.toResponse() = UserResponse(id, email, fullName, role, active, createdAt)
