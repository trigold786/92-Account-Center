import SwiftUI

@MainActor
class HomeViewModel: ObservableObject {
    @Published var isLoading = false
    @Published var errorMessage: String?
    private let authManager: AuthManager
    
    init(authManager: AuthManager = .shared) {
        self.authManager = authManager
    }
    
    var currentUser: User? { authManager.currentUser }
    
    func logout() async {
        await authManager.logout()
    }
}
