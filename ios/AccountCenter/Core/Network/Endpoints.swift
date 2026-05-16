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

    private static let subscriptionsBase = baseURL.appendingPathComponent("/api/v1/subscriptions")
    private static let accountBase = baseURL.appendingPathComponent("/api/v1/account")
    private static let creditsBase = baseURL.appendingPathComponent("/api/v1/credits")
    private static let referralBase = baseURL.appendingPathComponent("/api/v1/referral")
    private static let riskBase = baseURL.appendingPathComponent("/api/v1/risk")
    private static let pushBase = baseURL.appendingPathComponent("/api/v1/push")
    private static let dataBase = baseURL.appendingPathComponent("/api/v1/data")

    static func userSubscriptions(userId: Int64) -> URL {
        subscriptionsBase.appendingPathComponent("\(userId)")
    }

    static func userTier(userId: Int64) -> URL {
        accountBase.appendingPathComponent("\(userId)/tier")
    }

    static func creditAccount(userId: Int64) -> URL {
        creditsBase.appendingPathComponent("\(userId)/account")
    }

    static func transactions(userId: Int64, page: Int = 1, pageSize: Int = 20) -> URL {
        var components = URLComponents(url: creditsBase.appendingPathComponent("\(userId)/transactions"), resolvingAgainstBaseURL: false)!
        components.queryItems = [
            URLQueryItem(name: "page", value: "\(page)"),
            URLQueryItem(name: "page_size", value: "\(pageSize)")
        ]
        return components.url!
    }

    static let calculateDiscount = creditsBase.appendingPathComponent("calculate-discount")

    static func referralSummary(userId: Int64) -> URL {
        referralBase.appendingPathComponent("\(userId)/summary")
    }

    static let generateReferralLink = referralBase.appendingPathComponent("generate-link")

    static func riskHistory(userId: Int64) -> URL {
        riskBase.appendingPathComponent("history/\(userId)")
    }

    static func userDevices(userId: Int64) -> URL {
        pushBase.appendingPathComponent("user/\(userId)/devices")
    }

    static func rfmScore(userId: Int64) -> URL {
        dataBase.appendingPathComponent("rfm/\(userId)")
    }
}
