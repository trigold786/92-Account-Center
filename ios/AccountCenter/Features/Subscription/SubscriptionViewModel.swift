import SwiftUI

@MainActor
class SubscriptionViewModel: ObservableObject {
    @Published var subscriptions: [Subscription] = []
    @Published var currentTier: TierInfo?
    @Published var isLoading = false
    @Published var errorMessage: String?

    private let authManager: AuthManager

    init(authManager: AuthManager = .shared) {
        self.authManager = authManager
    }

    var userId: Int64? { authManager.userId }

    func load() async {
        guard let userId = userId else { return }
        isLoading = true
        errorMessage = nil

        do {
            async let subs = APIClient.shared.getUserSubscriptions(userId: userId)
            async let tier = APIClient.shared.getUserTier(userId: userId)
            (subscriptions, currentTier) = try await (subs, tier)
        } catch let error as APIError {
            errorMessage = error.errorDescription
        } catch {
            errorMessage = "加载失败，请稍后重试"
        }

        isLoading = false
    }

    var activeSubscription: Subscription? {
        subscriptions.first { $0.status == "active" }
    }

    func tierName(for level: Int) -> String {
        switch level {
        case 1: return "基础版"
        case 2: return "高级版"
        case 3: return "企业版"
        default: return "免费版"
        }
    }

    func tierColor(for level: Int) -> Color {
        switch level {
        case 1: return .blue
        case 2: return .orange
        case 3: return .purple
        default: return .gray
        }
    }
}
