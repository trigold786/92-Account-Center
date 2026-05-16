package com.accountcenter.model

import com.google.gson.annotations.SerializedName

data class LoginRequest(
    val credential: String,
    val password: String? = null,
    val code: String? = null,
    @SerializedName("magic_link") val magicLink: String? = null,
    @SerializedName("device_fingerprint_id") val deviceFingerprintId: String? = null
) {
    companion object {
        fun withPassword(phoneNumber: String, password: String) =
            LoginRequest(credential = phoneNumber, password = password)
        fun withCode(phoneNumber: String, code: String) =
            LoginRequest(credential = phoneNumber, code = code)
    }
}

data class RefreshTokenRequest(
    @SerializedName("refresh_token") val refreshToken: String
)

data class RegisterRequest(
    @SerializedName("phone_number") val phoneNumber: String,
    @SerializedName("account_id") val accountId: String,
    val password: String,
    @SerializedName("agree_to_terms") val agreeToTerms: Boolean,
    @SerializedName("referral_code") val referralCode: String? = null
)

data class SMSSendRequest(
    @SerializedName("phone_number") val phoneNumber: String
)

data class RegisterResponse(
    val id: Long,
    @SerializedName("phone_number") val phoneNumber: String,
    @SerializedName("account_id") val accountId: String,
    val message: String
)
