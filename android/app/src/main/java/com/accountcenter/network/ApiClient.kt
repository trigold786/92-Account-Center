package com.accountcenter.network

import com.accountcenter.model.ApiDataResponse
import com.accountcenter.model.CalculateDiscountRequest
import com.accountcenter.model.CreditAccount
import com.accountcenter.model.DeviceList
import com.accountcenter.model.DiscountInfo
import com.accountcenter.model.GenerateLinkRequest
import com.accountcenter.model.LoginRequest
import com.accountcenter.model.ReferralLinkData
import com.accountcenter.model.ReferralSummary
import com.accountcenter.model.RFMScore
import com.accountcenter.model.RefreshTokenRequest
import com.accountcenter.model.RegisterRequest
import com.accountcenter.model.RegisterResponse
import com.accountcenter.model.RiskHistoryData
import com.accountcenter.model.SMSSendRequest
import com.accountcenter.model.Subscription
import com.accountcenter.model.TierInfo
import com.accountcenter.model.Token
import com.accountcenter.model.TransactionList
import retrofit2.Response
import retrofit2.http.Body
import retrofit2.http.GET
import retrofit2.http.POST
import retrofit2.http.Path
import retrofit2.http.Query

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

    @GET("/api/v1/subscriptions/{user_id}")
    suspend fun getUserSubscriptions(@Path("user_id") userId: Long): Response<List<Subscription>>

    @GET("/api/v1/account/{user_id}/tier")
    suspend fun getUserTier(@Path("user_id") userId: Long): Response<TierInfo>

    @GET("/api/v1/credits/{user_id}/account")
    suspend fun getCreditAccount(@Path("user_id") userId: Long): Response<ApiDataResponse<CreditAccount>>

    @GET("/api/v1/credits/{user_id}/transactions")
    suspend fun getTransactions(
        @Path("user_id") userId: Long,
        @Query("page") page: Int = 1,
        @Query("page_size") pageSize: Int = 20
    ): Response<ApiDataResponse<TransactionList>>

    @GET("/api/v1/referral/{user_id}/summary")
    suspend fun getReferralSummary(@Path("user_id") userId: Long): Response<ApiDataResponse<ReferralSummary>>

    @POST("/api/v1/referral/generate-link")
    suspend fun generateReferralLink(@Body request: GenerateLinkRequest): Response<ApiDataResponse<ReferralLinkData>>

    @POST("/api/v1/credits/calculate-discount")
    suspend fun calculateDiscount(@Body request: CalculateDiscountRequest): Response<ApiDataResponse<DiscountInfo>>

    @GET("/api/v1/risk/history/{user_id}")
    suspend fun getRiskHistory(@Path("user_id") userId: Long): Response<ApiDataResponse<RiskHistoryData>>

    @GET("/api/v1/push/user/{user_id}/devices")
    suspend fun getUserDevices(@Path("user_id") userId: Long): Response<ApiDataResponse<DeviceList>>

    @GET("/api/v1/data/rfm/{user_id}")
    suspend fun getRFMScore(@Path("user_id") userId: Long): Response<RFMScore>
}
