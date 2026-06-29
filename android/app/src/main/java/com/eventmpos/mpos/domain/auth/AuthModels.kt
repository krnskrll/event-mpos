package com.eventmpos.mpos.domain.auth

/** Domain models — independent of network DTOs and framework types (Clean Architecture). */

enum class Role { ADMIN, BA, WAREHOUSE }

data class AuthUser(
    val id: String,
    val email: String,
    val fullName: String,
    val role: Role,
)

data class AssignedEvent(
    val id: String,
    val code: String,
    val name: String,
    val startDate: String?,
    val endDate: String?,
)

/** A typed result so the UI layer never has to touch exceptions/HTTP codes directly. */
sealed interface AuthResult<out T> {
    data class Success<T>(val data: T) : AuthResult<T>
    data class Failure(val message: String) : AuthResult<Nothing>
}
