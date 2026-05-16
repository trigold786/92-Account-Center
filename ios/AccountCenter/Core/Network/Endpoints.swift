import Foundation

enum Environment {
    case development
    case production
}

struct Endpoints {
    static let environment: Environment = .development
    
    static let baseURL: URL = {
        switch environment {
        case .development:
            return URL(string: "http://localhost:30300")!
        case .production:
            return URL(string: "https://api.accountcenter.com")!
        }
    }()
    
    static let login = baseURL.appendingPathComponent("/api/v1/auth/login")
    static let refresh = baseURL.appendingPathComponent("/api/v1/auth/refresh")
    static let logout = baseURL.appendingPathComponent("/api/v1/auth/logout")
    static let sendSMS = baseURL.appendingPathComponent("/api/v1/sms/send")
    static let register = baseURL.appendingPathComponent("/api/v1/account/register")
}
