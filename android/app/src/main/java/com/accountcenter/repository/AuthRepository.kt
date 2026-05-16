package com.accountcenter.repository

import com.accountcenter.model.LoginRequest
import com.accountcenter.model.RegisterRequest
import com.accountcenter.model.SMSSendRequest
import com.accountcenter.model.Token
import com.accountcenter.network.ApiClient
import com.accountcenter.network.ApiError
import com.accountcenter.storage.TokenManager
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.withContext
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class AuthRepository @Inject constructor(
    private val apiClient: ApiClient,
    private val tokenManager: TokenManager
) {
    val isAuthenticated: Flow<Boolean> = tokenManager.getTokenFlow().map { it != null }

    suspend fun login(request: LoginRequest): Result<Token> = withContext(Dispatchers.IO) {
        try {
            val response = apiClient.login(request)
            if (response.isSuccessful && response.body() != null) {
                val token = response.body()!!
                tokenManager.saveToken(token)
                Result.success(token)
            } else {
                Result.failure(ApiError.HttpError(response.code(), "登录失败"))
            }
        } catch (e: Exception) {
            Result.failure(e.toApiError())
        }
    }

    suspend fun sendSMS(request: SMSSendRequest): Result<Unit> = withContext(Dispatchers.IO) {
        try {
            val response = apiClient.sendSMS(request)
            if (response.isSuccessful) {
                Result.success(Unit)
            } else {
                Result.failure(ApiError.HttpError(response.code(), "发送失败"))
            }
        } catch (e: Exception) {
            Result.failure(e.toApiError())
        }
    }

    suspend fun register(request: RegisterRequest): Result<Unit> = withContext(Dispatchers.IO) {
        try {
            val response = apiClient.register(request)
            if (response.isSuccessful) {
                Result.success(Unit)
            } else {
                Result.failure(ApiError.HttpError(response.code(), "注册失败"))
            }
        } catch (e: Exception) {
            Result.failure(e.toApiError())
        }
    }

    suspend fun logout() = withContext(Dispatchers.IO) {
        try {
            apiClient.logout()
        } catch (e: Exception) {
        }
        tokenManager.clearTokens()
    }
}
