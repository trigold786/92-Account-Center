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
                            Text(viewModel.currentUser?.accountId ?? NSLocalizedString("user_placeholder", comment: ""))
                                .font(.custom("SpaceGrotesk-Bold", size: 18))
                                .foregroundColor(.textPrimary)
                            Text(NSLocalizedString("phone_unbound", comment: ""))
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
                        Text(NSLocalizedString("features_section", comment: ""))
                            .sectionTitle()
                            .padding(.horizontal, 4)
                            .padding(.bottom, 8)

                        VStack(spacing: 0) {
                            FeatureRow(icon: "cart", label: NSLocalizedString("feature_subscription", comment: ""), destination: SubscriptionView())
                            Divider().background(Color.divider).padding(.leading, 44)
                            FeatureRow(icon: "creditcard", label: NSLocalizedString("feature_credits", comment: ""), destination: CreditsView())
                            Divider().background(Color.divider).padding(.leading, 44)
                            FeatureRow(icon: "lock", label: NSLocalizedString("feature_security", comment: ""), destination: SecurityView())
                            Divider().background(Color.divider).padding(.leading, 44)
                            FeatureRow(icon: "info.circle", label: NSLocalizedString("feature_about", comment: ""), destination: AboutView())
                        }
                        .background(Color.bgCard)
                        .cornerRadius(16)
                    }

                    // Logout
                    Button(action: { Task { await viewModel.logout() } }) {
                        Text(NSLocalizedString("logout", comment: ""))
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
        .navigationTitle(NSLocalizedString("home_title", comment: ""))
        .navigationBarTitleDisplayMode(.inline)
        .toolbarBackground(Color.bgPrimary.opacity(0.9), for: .navigationBar)
        .task { await viewModel.loadRFM() }
        .preventScreenshot()
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
