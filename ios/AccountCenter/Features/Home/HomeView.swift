import SwiftUI

struct HomeView: View {
    @StateObject private var viewModel = HomeViewModel()

    var body: some View {
        ZStack {
            Color.bgPrimary.ignoresSafeArea()

            ScrollView {
                VStack(spacing: 12) {
                    // User Card
                    HStack(spacing: 16) {
                        ZStack {
                            Circle()
                                .fill(Color.brandGradient)
                                .frame(width: 64, height: 64)
                            Text(viewModel.currentUser?.accountId.prefix(1).uppercased() ?? "U")
                                .font(.custom("SpaceGrotesk-Bold", size: 24))
                                .foregroundColor(.white)
                        }
                        VStack(alignment: .leading, spacing: 4) {
                            Text(viewModel.currentUser?.accountId ?? "用户")
                                .font(.custom("SpaceGrotesk-Bold", size: 18))
                                .foregroundColor(.textPrimary)
                            Text("未绑定手机号")
                                .font(.custom("Inter-Regular", size: 13))
                                .foregroundColor(.textSecondary)
                            HStack(spacing: 4) {
                                Circle()
                                    .fill(Color.brandSecondary)
                                    .frame(width: 6, height: 6)
                                Text("Lv.2")
                                    .font(.custom("SpaceGrotesk-Semibold", size: 11))
                                    .foregroundColor(.brandSecondary)
                            }
                        }
                        Spacer()
                    }
                    .cardStyle()

                    // RFM Card
                    if let rfm = viewModel.rfmScore {
                        HStack(spacing: 12) {
                            Text("\u{1F3AF}")
                                .font(.title2)
                            VStack(alignment: .leading, spacing: 2) {
                                Text(rfm.rfmSegmentCn)
                                    .font(.custom("SpaceGrotesk-Semibold", size: 14))
                                    .foregroundColor(.textPrimary)
                                Text("RFM \(rfm.rfmSegment)")
                                    .font(.custom("Inter-Regular", size: 12))
                                    .foregroundColor(.textSecondary)
                            }
                            Spacer()
                        }
                        .cardStyle()
                        .overlay(
                            RoundedRectangle(cornerRadius: 16)
                                .stroke(Color.brandGradient, lineWidth: 1)
                                .opacity(0.3)
                        )
                    }

                    // Features
                    VStack(alignment: .leading, spacing: 0) {
                        Text("功能")
                            .sectionTitle()
                            .padding(.horizontal, 4)
                            .padding(.bottom, 8)

                        VStack(spacing: 0) {
                            FeatureRow(icon: "cart", label: "订阅管理", destination: SubscriptionView())
                            Divider().background(Color.divider).padding(.leading, 44)
                            FeatureRow(icon: "creditcard", label: "积分中心", destination: CreditsView())
                            Divider().background(Color.divider).padding(.leading, 44)
                            FeatureRow(icon: "lock", label: "安全设置", destination: SecurityView())
                            Divider().background(Color.divider).padding(.leading, 44)
                            FeatureRow(icon: "info.circle", label: "关于", destination: AboutView())
                        }
                        .background(Color.bgCard)
                        .cornerRadius(16)
                    }

                    // Logout
                    Button(action: { Task { await viewModel.logout() } }) {
                        Text("退出登录")
                            .font(.custom("Inter-Regular", size: 15))
                            .foregroundColor(.danger)
                            .frame(maxWidth: .infinity)
                            .frame(height: 52)
                            .background(Color.bgCard)
                            .cornerRadius(12)
                    }
                    .padding(.top, 8)
                }
                .padding(16)
            }
        }
        .navigationTitle("用户中心")
        .navigationBarTitleDisplayMode(.inline)
        .toolbarBackground(Color.bgPrimary.opacity(0.9), for: .navigationBar)
        .task { await viewModel.loadRFM() }
    }
}

private struct FeatureRow<Destination: View>: View {
    let icon: String
    let label: String
    let destination: Destination

    var body: some View {
        NavigationLink(destination: destination) {
            HStack(spacing: 12) {
                Image(systemName: icon)
                    .font(.system(size: 18))
                    .foregroundColor(.brandPrimary)
                    .frame(width: 24)
                Text(label)
                    .font(.custom("Inter-Regular", size: 15))
                    .foregroundColor(.textPrimary)
                Spacer()
                Image(systemName: "chevron.right")
                    .font(.system(size: 12, weight: .semibold))
                    .foregroundColor(.textSecondary)
            }
            .frame(height: 56)
            .padding(.horizontal, 16)
        }
    }
}
