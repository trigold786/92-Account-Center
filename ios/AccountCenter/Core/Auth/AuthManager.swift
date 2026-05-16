import Foundation
import SwiftUI

@MainActor
class AuthManager: ObservableObject {
    static let shared = AuthManager()
    
    private init() {
        loadToken()
    }
    
    var isAuthenticated: Bool {
        accessToken != nil
    }
    
    @Published var currentUser: User?
    
    @Published private(set) var accessToken: String?
    @Published private(set) var refreshToken: String?
    @Published private(set) var userId: Int64?
    
    private func loadToken() {
        accessToken = TokenManager.shared.getAccessToken()
        refreshToken = TokenManager.shared.getRefreshToken()
        currentUser = accessToken.map { _ in
            User(
                id: TokenManager.shared.getUserId() ?? 0,
                phoneNumber: nil,
                accountId: TokenManager.shared.getAccountId() ?? "",
                email: nil,
                mfaEnabled: nil,
                createdAt: nil,
                updatedAt: nil
            )
        }
    }
    
    func login(token: Token) {
        TokenManager.shared.save(token: token)
        accessToken = token.accessToken
        refreshToken = token.refreshToken
        userId = token.userId
        
        currentUser = User(
            id: token.userId,
            phoneNumber: nil,
            accountId: token.accountId,
            email: nil,
            mfaEnabled: nil,
            createdAt: nil,
            updatedAt: nil
        )
    }
    
    func logout() async {
        if let accessToken = accessToken {
            try? await APIClient.shared.logout(accessToken: accessToken)
        }
        
        TokenManager.shared.clear()
        accessToken = nil
        refreshToken = nil
        userId = nil
        currentUser = nil
    }
    
    func refreshIfNeeded() async -> Bool {
        guard let currentRefreshToken = refreshToken else {
            return false
        }
        
        do {
            let newToken = try await APIClient.shared.refresh(refreshToken: currentRefreshToken)
            login(token: newToken)
            return true
        } catch {
            await logout()
            return false
        }
    }
}
