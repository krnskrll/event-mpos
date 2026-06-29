package com.eventmpos.mpos.data.auth

import com.eventmpos.mpos.domain.auth.AssignedEvent
import com.eventmpos.mpos.domain.auth.AuthUser
import com.eventmpos.mpos.domain.auth.Role
import kotlinx.serialization.Serializable

@Serializable
data class LoginRequestDto(val email: String, val password: String)

@Serializable
data class RefreshRequestDto(val refreshToken: String)

@Serializable
data class UserSummaryDto(
    val id: String,
    val email: String,
    val fullName: String,
    val role: String,
)

@Serializable
data class TokenResponseDto(
    val accessToken: String,
    val refreshToken: String,
    val expiresInSeconds: Long,
    val user: UserSummaryDto,
)

@Serializable
data class EventResponseDto(
    val id: String,
    val code: String,
    val name: String,
    val startDate: String? = null,
    val endDate: String? = null,
    val active: Boolean = true,
)

fun UserSummaryDto.toDomain() = AuthUser(id, email, fullName, Role.valueOf(role))
fun EventResponseDto.toDomain() = AssignedEvent(id, code, name, startDate, endDate)
