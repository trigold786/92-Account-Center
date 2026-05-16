import SwiftUI

struct RootView: View {
    @StateObject private var authManager = AuthManager.shared
    
    var body: some View {
        NavigationStack {
            if authManager.isAuthenticated {
                HomeView()
                    .environmentObject(authManager)
            } else {
                LoginView()
                    .environmentObject(authManager)
            }
        }
    }
}

#Preview {
    RootView()
}
