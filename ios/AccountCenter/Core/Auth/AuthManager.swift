import Foundation
import SwiftUI

@MainActor
@Observable
class AuthManager {
    static let shared = AuthManager()
    
    private init() {
        loadToken()
    }
    
    var isAuthenticated: Bool {
        accessToken != nil
    }
    
    var currentUser: User?
    
    private(set) var accessToken: String?
    private(set) var refreshToken: String?
    private(set) var userId: Int64?
    
    private func loadToken() {
        accessToken = TokenManager.shared.getAccessToken()
        refreshToken = TokenManager.shared.getRefreshToken()
    }
    
    func login(token: Token) {
        TokenManager.shared.save(token: token)
        accessToken = token.accessToken
        refreshToken = token.refreshToken
        userId = token.userId
        
        currentUser = User(
            id: token.userId,
            phoneNumber: "138****1234",
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
    
    func refreshIfNeeded() async {
        guard let refreshToken = refreshToken else {
            return
        }
        
        do {
            let newToken = try await APIClient.shared.refresh(refreshToken: refreshToken)
            login(token: newToken)
        } catch {
            await logout()
        }
    }
}
