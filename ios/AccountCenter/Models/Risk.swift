import Foundation

struct RiskEvent: Codable, Identifiable {
    let riskEventId: String
    let eventType: String
    let riskScore: Int
    let riskLevel: String
    let details: [String: String]?
    let ipAddress: String?
    let createdAt: String

    var id: String { riskEventId }

    enum CodingKeys: String, CodingKey {
        case riskEventId = "risk_event_id"
        case eventType = "event_type"
        case riskScore = "risk_score"
        case riskLevel = "risk_level"
        case details
        case ipAddress = "ip_address"
        case createdAt = "created_at"
    }
}

struct RiskHistoryData: Codable {
    let events: [RiskEvent]
    let limit: Int?
}

struct PushDevice: Codable, Identifiable {
    let id: String
    let deviceToken: String
    let platform: String
    let deviceName: String?
    let isActive: Bool
    let lastActiveAt: String?

    enum CodingKeys: String, CodingKey {
        case id
        case deviceToken = "device_token"
        case platform
        case deviceName = "device_name"
        case isActive = "is_active"
        case lastActiveAt = "last_active_at"
    }
}

struct DeviceList: Codable {
    let devices: [PushDevice]
}
