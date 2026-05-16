package com.accountcenter.network

import com.accountcenter.model.LoginRequest
import com.accountcenter.model.RefreshTokenRequest
import com.accountcenter.model.RegisterRequest
import com.accountcenter.model.RegisterResponse
import com.accountcenter.model.SMSSendRequest
import com.accountcenter.model.Token
import retrofit2.Response
import retrofit2.http.Body
import retrofit2.http.POST

interface ApiClient {
    @POST("/api/v1/auth/login")
    suspend fun login(@Body request: LoginRequest): Response<Token>

    @POST("/api/v1/auth/refresh")
    suspend fun refresh(@Body request: RefreshTokenRequest): Response<Token>

    @POST("/api/v1/auth/logout")
    suspend fun logout(): Response<Unit>

    @POST("/api/v1/sms/send")
    suspend fun sendSMS(@Body request: SMSSendRequest): Response<Map<String, String>>

    @POST("/api/v1/account/register")
    suspend fun register(@Body request: RegisterRequest): Response<RegisterResponse>
}
