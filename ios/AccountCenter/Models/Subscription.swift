import Foundation

struct Subscription: Codable, Identifiable {
    let id: Int64
    let userId: Int64
    let tierLevel: Int
    let startTime: String
    let endTime: String
    let status: String
    let price: Double
    let paymentMethod: String?
    let orderId: String?

    enum CodingKeys: String, CodingKey {
        case id
        case userId = "user_id"
        case tierLevel = "tier_level"
        case startTime = "start_time"
        case endTime = "end_time"
        case status
        case price
        case paymentMethod = "payment_method"
        case orderId = "order_id"
    }
}

struct TierInfo: Codable {
    let userId: Int64
    let identityTier: Int

    enum CodingKeys: String, CodingKey {
        case userId = "user_id"
        case identityTier = "identity_tier"
    }
}
