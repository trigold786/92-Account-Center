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
    override fun authenticate(route: Route?, response: Response): Request? {
        if (response.request.header("Authorization") != null) {
            return null
        }

        return runBlocking {
            val refreshToken = tokenManager.getRefreshToken()
            if (refreshToken != null) {
                try {
                    val newToken = apiClient.refresh(RefreshTokenRequest(refreshToken))
                    if (newToken.isSuccessful && newToken.body() != null) {
                        tokenManager.saveToken(newToken.body()!!)
                        return@runBlocking response.request.newBuilder()
                            .header("Authorization", "Bearer ${newToken.body()!!.accessToken}")
                            .build()
                    }
                } catch (e: Exception) {
                    tokenManager.clearTokens()
                }
            }

            tokenManager.clearTokens()
            null
        }
    }
}
