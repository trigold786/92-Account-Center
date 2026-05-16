package com.accountcenter.model

import com.google.gson.annotations.SerializedName

data class Token(
    @SerializedName("access_token") val accessToken: String,
    @SerializedName("refresh_token") val refreshToken: String,
    @SerializedName("expires_in") val expiresIn: Long,
    @SerializedName("user_id") val userId: Long,
    @SerializedName("account_id") val accountId: String
)
