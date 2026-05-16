import SwiftUI

@MainActor
class HomeViewModel: ObservableObject {
    @Published var isLoading = false
    @Published var errorMessage: String?
    @Published var rfmScore: RFMScore?
    private let authManager: AuthManager
    
    init(authManager: AuthManager = .shared) {
        self.authManager = authManager
    }
    
    var currentUser: User? { authManager.currentUser }
    
    func logout() async {
        await authManager.logout()
    }
    
    func loadRFM() async {
        guard let userId = authManager.currentUser?.id else { return }
        rfmScore = try? await APIClient.shared.getRFMScore(userId: userId)
    }
}
