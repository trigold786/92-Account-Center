package com.accountcenter.model

import com.google.gson.annotations.SerializedName

data class User(
    val id: Long,
    @SerializedName("phone_number") val phoneNumber: String? = null,
    @SerializedName("account_id") val accountId: String,
    val email: String? = null,
    @SerializedName("mfa_enabled") val mfaEnabled: Boolean? = null
)
