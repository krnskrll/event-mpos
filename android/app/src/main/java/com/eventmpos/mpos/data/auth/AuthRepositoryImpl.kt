package com.eventmpos.mpos.data.auth

import com.eventmpos.mpos.core.secure.SecureTokenStorage
import com.eventmpos.mpos.domain.auth.AssignedEvent
import com.eventmpos.mpos.domain.auth.AuthRepository
import com.eventmpos.mpos.domain.auth.AuthResult
import com.eventmpos.mpos.domain.auth.AuthUser
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import retrofit2.HttpException
import java.io.IOException
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class AuthRepositoryImpl @Inject constructor(
    private val api: AuthApi,
    private val tokenStorage: SecureTokenStorage,
) : AuthRepository {

    override fun isLoggedIn(): Boolean = tokenStorage.isLoggedIn

    override suspend fun login(email: String, password: String): AuthResult<AuthUser> =
        withContext(Dispatchers.IO) {
            try {
                val res = api.login(LoginRequestDto(email.trim(), password))
                tokenStorage.saveTokens(res.accessToken, res.refreshToken)
                AuthResult.Success(res.user.toDomain())
            } catch (e: HttpException) {
                AuthResult.Failure(
                    if (e.code() == 401) "Invalid email or password" else "Login failed (${e.code()})",
                )
            } catch (e: IOException) {
                AuthResult.Failure("Can't reach the server. Check the venue network.")
            }
        }

    override suspend fun getAssignedEvent(): AuthResult<AssignedEvent> =
        withContext(Dispatchers.IO) {
            try {
                AuthResult.Success(api.getMyEvent().toDomain())
            } catch (e: HttpException) {
                AuthResult.Failure(
                    if (e.code() == 404) "No event is assigned to your account yet."
                    else "Couldn't load your event (${e.code()})",
                )
            } catch (e: IOException) {
                AuthResult.Failure("Can't reach the server. Check the venue network.")
            }
        }

    override suspend fun logout() {
        withContext(Dispatchers.IO) {
            runCatching { api.logout() } // best-effort server-side revocation
            tokenStorage.clear()
        }
    }
}
