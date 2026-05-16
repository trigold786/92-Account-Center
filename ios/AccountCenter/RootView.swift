import SwiftUI

struct RootView: View {
    @ObservedObject private var authManager = AuthManager.shared
    
    var body: some View {
        NavigationStack {
            if authManager.isAuthenticated {
                HomeView()
            } else {
                LoginView()
            }
        }
    }
}

#Preview {
    RootView()
}
