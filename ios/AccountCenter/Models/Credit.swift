import Foundation

struct CreditAccount: Codable {
    let userId: Int64
    let balance: Double
    let status: String

    enum CodingKeys: String, CodingKey {
        case userId = "user_id"
        case balance
        case status
    }
}

struct Transaction: Codable, Identifiable {
    let id: Int64
    let creditAccountId: Int64
    let type: String
    let amount: Double
    let referenceId: String?
    let details: String?
    let status: String
    let createdAt: String

    enum CodingKeys: String, CodingKey {
        case id
        case creditAccountId = "credit_account_id"
        case type
        case amount
        case referenceId = "reference_id"
        case details
        case status
        case createdAt = "created_at"
    }
}

struct TransactionList: Codable {
    let transactions: [Transaction]
    let total: Int
    let page: Int
    let pageSize: Int

    enum CodingKeys: String, CodingKey {
        case transactions
        case total
        case page
        case pageSize = "page_size"
    }
}

struct ReferralSummary: Codable {
    let totalReferees: Int
    let totalEarned: Double
    let activeReferees: Int

    enum CodingKeys: String, CodingKey {
        case totalReferees = "total_referees"
        case totalEarned = "total_earned"
        case activeReferees = "active_referees"
    }
}

struct CalculateDiscountRequest: Codable {
    let userId: String
    let subscriptionPrice: Double

    enum CodingKeys: String, CodingKey {
        case userId = "user_id"
        case subscriptionPrice = "subscription_price"
    }
}

struct DiscountInfo: Codable {
    let availableBalance: Double
    let maxDiscount: Double
    let remainingToPay: Double

    enum CodingKeys: String, CodingKey {
        case availableBalance = "available_balance"
        case maxDiscount = "max_discount"
        case remainingToPay = "remaining_to_pay"
    }
}

struct GenerateLinkRequest: Codable {
    let userId: String

    enum CodingKeys: String, CodingKey {
        case userId = "user_id"
    }
}

struct ReferralLinkData: Codable {
    let referralCode: String
    let referralLink: String

    enum CodingKeys: String, CodingKey {
        case referralCode = "referral_code"
        case referralLink = "referral_link"
    }
}
