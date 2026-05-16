package com.accountcenter.network

import com.accountcenter.model.RefreshTokenRequest
import com.accountcenter.storage.TokenManager
import kotlinx.coroutines.runBlocking
import okhttp3.Authenticator
import okhttp3.Request
import okhttp3.Response
import okhttp3.Route
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class TokenAuthenticator @Inject constructor(
    private val tokenManager: TokenManager,
    private val apiClient: ApiClient
) : Authenticator {
    private val lock = Any()

    override fun authenticate(route: Route?, response: Response): Request? {
        if (response.request.url.encodedPath.contains("/auth/refresh")) return null
        if (route == null) return null

        synchronized(lock) {
            val refreshToken = runBlocking { tokenManager.getRefreshToken() } ?: return null

            return runBlocking {
                try {
                    val newTokenResponse = apiClient.refresh(RefreshTokenRequest(refreshToken))
                    if (newTokenResponse.isSuccessful && newTokenResponse.body() != null) {
                        val newToken = newTokenResponse.body()!!
                        tokenManager.saveToken(newToken)
                        response.request.newBuilder()
                            .header("Authorization", "Bearer ${newToken.accessToken}")
                            .build()
                    } else {
                        tokenManager.clearTokens()
                        null
                    }
                } catch (e: Exception) {
                    null
                }
            }
        }
    }
}
