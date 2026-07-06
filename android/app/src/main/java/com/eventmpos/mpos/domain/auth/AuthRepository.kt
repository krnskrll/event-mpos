package com.eventmpos.mpos.domain.auth

interface AuthRepository {
    /** True if valid tokens are already stored (used to skip the login screen on relaunch). */
    fun isLoggedIn(): Boolean

    suspend fun login(email: String, password: String): AuthResult<AuthUser>

    /** The active event assigned to the logged-in BA. */
    suspend fun getAssignedEvent(): AuthResult<AssignedEvent>

    suspend fun logout()
}
