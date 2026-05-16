import SwiftUI

@MainActor
class HomeViewModel: ObservableObject {
    @Published var isLoading = false
    @Published var errorMessage: String?
    @EnvironmentObject var authManager: AuthManager
    
    func logout() async {
        await authManager.logout()
    }
}
