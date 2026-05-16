import Foundation

struct Token: Codable {
    let accessToken: String
    let refreshToken: String
    let expiresIn: Int64
    let userId: Int64
    let accountId: String
    
    enum CodingKeys: String, CodingKey {
        case accessToken = "access_token"
        case refreshToken = "refresh_token"
        case expiresIn = "expires_in"
        case userId = "user_id"
        case accountId = "account_id"
    }
}
