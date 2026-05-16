package com.accountcenter.model

import com.google.gson.annotations.SerializedName

data class ApiDataResponse<T>(
    val code: Int,
    val message: String? = null,
    val data: T? = null
)

data class Subscription(
    val id: Long,
    @SerializedName("user_id") val userId: Long,
    @SerializedName("tier_level") val tierLevel: Int,
    @SerializedName("start_time") val startTime: String,
    @SerializedName("end_time") val endTime: String,
    val status: String,
    val price: Double,
    @SerializedName("payment_method") val paymentMethod: String? = null,
    @SerializedName("order_id") val orderId: String? = null
)

data class TierInfo(
    @SerializedName("user_id") val userId: Long,
    @SerializedName("identity_tier") val identityTier: Int
)

data class CreditAccount(
    @SerializedName("user_id") val userId: Long,
    val balance: Double,
    val status: String
)

data class Transaction(
    val id: Long,
    @SerializedName("credit_account_id") val creditAccountId: Long,
    val type: String,
    val amount: Double,
    @SerializedName("reference_id") val referenceId: String? = null,
    val details: String? = null,
    val status: String,
    @SerializedName("created_at") val createdAt: String
)

data class TransactionList(
    val transactions: List<Transaction>,
    val total: Int,
    val page: Int,
    @SerializedName("page_size") val pageSize: Int
)

data class ReferralSummary(
    @SerializedName("total_referees") val totalReferees: Int,
    @SerializedName("total_earned") val totalEarned: Double,
    @SerializedName("active_referees") val activeReferees: Int
)

data class CalculateDiscountRequest(
    @SerializedName("user_id") val userId: String,
    @SerializedName("subscription_price") val subscriptionPrice: Double
)

data class DiscountInfo(
    @SerializedName("available_balance") val availableBalance: Double,
    @SerializedName("max_discount") val maxDiscount: Double,
    @SerializedName("remaining_to_pay") val remainingToPay: Double
)

data class RiskEvent(
    @SerializedName("risk_event_id") val riskEventId: String,
    @SerializedName("event_type") val eventType: String,
    @SerializedName("risk_score") val riskScore: Int,
    @SerializedName("risk_level") val riskLevel: String,
    val details: Map<String, String>? = null,
    @SerializedName("ip_address") val ipAddress: String? = null,
    @SerializedName("created_at") val createdAt: String
)

data class RiskHistoryData(
    val events: List<RiskEvent>,
    val limit: Int? = null
)

data class PushDevice(
    val id: String,
    @SerializedName("device_token") val deviceToken: String,
    val platform: String,
    @SerializedName("device_name") val deviceName: String? = null,
    @SerializedName("is_active") val isActive: Boolean,
    @SerializedName("last_active_at") val lastActiveAt: String? = null
)

data class DeviceList(
    val devices: List<PushDevice>
)

data class GenerateLinkRequest(
    @SerializedName("user_id") val userId: String
)

data class ReferralLinkData(
    @SerializedName("referral_code") val referralCode: String,
    @SerializedName("referral_link") val referralLink: String
)

data class RFMScore(
    @SerializedName("user_id") val userId: Long,
    @SerializedName("recency_score") val recencyScore: Int,
    @SerializedName("frequency_score") val frequencyScore: Int,
    @SerializedName("monetary_score") val monetaryScore: Int,
    @SerializedName("rfm_segment") val rfmSegment: String,
    @SerializedName("rfm_segment_cn") val rfmSegmentCn: String,
    @SerializedName("last_subscription_at") val lastSubscriptionAt: String? = null,
    @SerializedName("total_subscriptions") val totalSubscriptions: Int = 0,
    @SerializedName("total_spent") val totalSpent: Double = 0.0
)
