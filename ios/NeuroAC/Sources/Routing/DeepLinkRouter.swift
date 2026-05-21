import Foundation

enum DeepLinkDestination {
    case subscription(tier: String?)
    case referralRegister(inviterId: String)
    case credits
    case home
}

class DeepLinkRouter {
    static let shared = DeepLinkRouter()

    private let supportedHosts = ["subscribe", "register", "credits", "referral"]
    private let universalLinkHost = "ac.neuro.ai"

    func route(url: URL) -> DeepLinkDestination? {
        if url.scheme == "neuro" {
            return parseCustomScheme(url)
        } else if url.scheme == "https", url.host == universalLinkHost {
            return parseUniversalLink(url)
        }
        return nil
    }

    private func parseCustomScheme(_ url: URL) -> DeepLinkDestination? {
        guard let host = url.host else { return nil }
        let params = extractQueryParams(url)

        switch host {
        case "subscribe":
            let tier = params["tier"]
            return .subscription(tier: tier)
        case "register":
            if let inviterId = params["inviter_id"] {
                return .referralRegister(inviterId: inviterId)
            }
            return .home
        case "referral":
            if let inviterId = params["inviter_id"] {
                return .referralRegister(inviterId: inviterId)
            }
            return .home
        case "credits":
            return .credits
        default:
            return .home
        }
    }

    private func parseUniversalLink(_ url: URL) -> DeepLinkDestination? {
        let pathComponents = url.pathComponents.filter { $0 != "/" }

        guard pathComponents.count >= 1 else { return .home }
        let params = extractQueryParams(url)

        switch pathComponents[0] {
        case "subscribe":
            let tier = params["tier"] ?? (pathComponents.count > 1 ? pathComponents[1] : nil)
            return .subscription(tier: tier)
        case "register":
            if let inviterId = params["inviter_id"] {
                return .referralRegister(inviterId: inviterId)
            }
            return .home
        case "referral":
            if let inviterId = params["inviter_id"] {
                return .referralRegister(inviterId: inviterId)
            }
            return .home
        case "credits":
            return .credits
        default:
            return .home
        }
    }

    private func extractQueryParams(_ url: URL) -> [String: String] {
        guard let components = URLComponents(url: url, resolvingAgainstBaseURL: false),
              let queryItems = components.queryItems else { return [:] }
        return queryItems.reduce(into: [String: String]()) { result, item in
            if let value = item.value {
                result[item.name] = value
            }
        }
    }
}
