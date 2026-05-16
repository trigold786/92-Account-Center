import Foundation

struct RFMScore: Codable {
    let userId: Int64
    let recencyScore: Int
    let frequencyScore: Int
    let monetaryScore: Int
    let rfmSegment: String
    let rfmSegmentCn: String
    let lastSubscriptionAt: String?
    let totalSubscriptions: Int
    let totalSpent: Double

    enum CodingKeys: String, CodingKey {
        case userId = "user_id"
        case recencyScore = "recency_score"
        case frequencyScore = "frequency_score"
        case monetaryScore = "monetary_score"
        case rfmSegment = "rfm_segment"
        case rfmSegmentCn = "rfm_segment_cn"
        case lastSubscriptionAt = "last_subscription_at"
        case totalSubscriptions = "total_subscriptions"
        case totalSpent = "total_spent"
    }
}
