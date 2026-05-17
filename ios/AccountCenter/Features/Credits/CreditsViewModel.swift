import SwiftUI

@MainActor
class CreditsViewModel: ObservableObject {
    @Published var account: CreditAccount?
    @Published var transactions: [Transaction] = []
    @Published var referral: ReferralSummary?
    @Published var referralLink: String?
    @Published var isLoading = false
    @Published var isLoadingMore = false
    @Published var errorMessage: String?
    @Published var hasMore = true

    private let authManager: AuthManager
    private var currentPage = 1

    init(authManager: AuthManager = .shared) {
        self.authManager = authManager
    }

    var userId: Int64? { authManager.userId }

    func load() async {
        guard let userId = userId else { return }
        isLoading = true
        errorMessage = nil
        currentPage = 1
        hasMore = true

        do {
            async let acct = APIClient.shared.getCreditAccount(userId: userId)
            async let txn = APIClient.shared.getTransactions(userId: userId, page: 1)
            async let ref = APIClient.shared.getReferralSummary(userId: userId)
            let (loadedAccount, loadedTxns, loadedReferral) = try await (acct, txn, ref)
            account = loadedAccount
            transactions = loadedTxns.transactions
            hasMore = loadedTxns.transactions.count >= loadedTxns.pageSize
            referral = loadedReferral
        } catch let error as APIError {
            errorMessage = error.errorDescription
        } catch {
            errorMessage = "加载失败，请稍后重试"
        }

        isLoading = false
    }

    func loadMore() async {
        guard let userId = userId, !isLoadingMore, hasMore else { return }
        isLoadingMore = true
        currentPage += 1

        do {
            let txn = try await APIClient.shared.getTransactions(userId: userId, page: currentPage)
            transactions.append(contentsOf: txn.transactions)
            hasMore = txn.transactions.count >= txn.pageSize
        } catch {
            currentPage -= 1
        }

        isLoadingMore = false
    }

    func earnCredits(amount: Int, reason: String) async {
        do {
            let _ = try await APIClient.shared.request("/api/v1/credits/earn", method: "POST",
                body: ["user_id": userId, "amount": amount, "reason": reason])
            await loadCredits()
        } catch {
            errorMessage = "签到失败"
        }
    }

    func generateLink() async {
        guard let userId = userId else { return }
        do {
            let data = try await APIClient.shared.generateReferralLink(userId: userId)
            referralLink = data.referralLink
            UIPasteboard.general.string = data.referralLink
        } catch let error as APIError {
            errorMessage = error.errorDescription
        } catch {
            errorMessage = "生成推荐链接失败"
        }
    }

    func transactionTypeIcon(_ type: String) -> String {
        switch type {
        case "earn", "referral_bonus": return "plus.circle.fill"
        case "consume", "subscription_payment": return "minus.circle.fill"
        case "refund": return "arrow.uturn.left.circle.fill"
        default: return "circle.fill"
        }
    }

    func transactionTypeColor(_ type: String) -> Color {
        switch type {
        case "earn", "referral_bonus", "refund": return .green
        case "consume", "subscription_payment": return .red
        default: return .gray
        }
    }
}
