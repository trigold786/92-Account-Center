import Foundation

struct LoginRequest: Codable {
    let credential: String
    let password: String?
    let code: String?
    let magicLink: String?
    let deviceFingerprintId: String?
    
    enum CodingKeys: String, CodingKey {
        case credential
        case password
        case code
        case magicLink = "magic_link"
        case deviceFingerprintId = "device_fingerprint_id"
    }
    
    init(phoneNumber: String, password: String) {
        self.credential = phoneNumber
        self.password = password
        self.code = nil
        self.magicLink = nil
        self.deviceFingerprintId = nil
    }
    
    init(phoneNumber: String, code: String) {
        self.credential = phoneNumber
        self.password = nil
        self.code = code
        self.magicLink = nil
        self.deviceFingerprintId = nil
    }
}

struct RefreshTokenRequest: Codable {
    let refreshToken: String
    
    enum CodingKeys: String, CodingKey {
        case refreshToken = "refresh_token"
    }
}

struct RegisterRequest: Codable {
    let phoneNumber: String
    let accountId: String
    let password: String
    let agreeToTerms: Bool
    let referralCode: String?
    
    enum CodingKeys: String, CodingKey {
        case phoneNumber = "phone_number"
        case accountId = "account_id"
        case password
        case agreeToTerms = "agree_to_terms"
        case referralCode = "referral_code"
    }
}

struct SMSSendRequest: Codable {
    let phoneNumber: String
    let templateCode: String?
    let params: [String: String]?
    
    enum CodingKeys: String, CodingKey {
        case phoneNumber = "phone_number"
        case templateCode = "template_code"
        case params
    }
    
    init(phoneNumber: String) {
        self.phoneNumber = phoneNumber
        self.templateCode = nil
        self.params = nil
    }
}

struct RegisterResponse: Codable {
    let id: Int64
    let phoneNumber: String
    let accountId: String
    let message: String
}
