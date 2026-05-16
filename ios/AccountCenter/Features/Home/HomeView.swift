import SwiftUI

struct HomeView: View {
    @StateObject private var viewModel = HomeViewModel()
    
    var body: some View {
        List {
            Section {
                HStack(spacing: 16) {
                    Circle()
                        .fill(Color.blue.opacity(0.2))
                        .frame(width: 64, height: 64)
                        .overlay(
                            Text(viewModel.currentUser?.accountId.prefix(1).uppercased() ?? "U")
                                .font(.title)
                                .fontWeight(.bold)
                                .foregroundColor(.blue)
                        )
                    
                    VStack(alignment: .leading, spacing: 4) {
                        Text(viewModel.currentUser?.accountId ?? "用户")
                            .font(.headline)
                        Text("未绑定手机号")
                            .font(.subheadline)
                            .foregroundStyle(.secondary)
                    }
                }
                .padding(.vertical, 8)
            }
            
            Section("功能") {
                NavigationLink(destination: EmptyView()) {
                    Label("订阅管理", systemImage: "cart")
                }
                NavigationLink(destination: EmptyView()) {
                    Label("积分中心", systemImage: "creditcard")
                }
                NavigationLink(destination: EmptyView()) {
                    Label("安全设置", systemImage: "lock")
                }
                NavigationLink(destination: EmptyView()) {
                    Label("关于", systemImage: "info.circle")
                }
            }
            
            Section {
                Button(action: {
                    Task {
                        await viewModel.logout()
                    }
                }) {
                    Text("退出登录")
                        .foregroundColor(.red)
                        .frame(maxWidth: .infinity)
                }
            }
        }
        .navigationTitle("用户中心")
        .navigationBarTitleDisplayMode(.inline)
    }
}

struct HomeView_Previews: PreviewProvider {
    static var previews: some View {
        HomeView()
    }
}
