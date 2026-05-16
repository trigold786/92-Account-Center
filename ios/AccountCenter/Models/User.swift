import Foundation

struct User: Codable, Identifiable {
    let id: Int64
    let phoneNumber: String
    let accountId: String
    let email: String?
    let mfaEnabled: Bool?
    let createdAt: Date?
    let updatedAt: Date?
    
    enum CodingKeys: String, CodingKey {
        case id
        case phoneNumber = "phone_number"
        case accountId = "account_id"
        case email
        case mfaEnabled = "mfa_enabled"
        case createdAt = "created_at"
        case updatedAt = "updated_at"
    }
}
