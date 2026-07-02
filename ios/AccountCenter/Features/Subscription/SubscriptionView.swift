import SwiftUI

struct SubscriptionView: View {
    @StateObject private var viewModel = SubscriptionViewModel()

    var body: some View {
        ZStack {
            Color.bgPrimary.ignoresSafeArea()

            ScrollView {
                VStack(spacing: 12) {
                    if viewModel.isLoading && viewModel.subscriptions.isEmpty {
                        ProgressView().padding()
                    }

                    if let errorMessage = viewModel.errorMessage {
                        Text(errorMessage)
                            .font(.custom("Inter-Regular", size: 12))
                            .foregroundColor(.danger)
                    }

                    // Tier badge
                    if let tier = viewModel.currentTier {
                        VStack(alignment: .leading, spacing: 8) {
                            HStack {
                                Circle()
                                    .fill(viewModel.tierColor(for: tier.identityTier))
                                    .frame(width: 12, height: 12)
                                Text(viewModel.tierName(for: tier.identityTier))
                                    .font(.custom("SpaceGrotesk-Semibold", size: 16))
                                    .foregroundColor(.textPrimary)
                                Spacer()
                                Text("Lv.\(tier.identityTier)")
                                    .font(.custom("SpaceGrotesk-Semibold", size: 12))
                                    .foregroundColor(.brandSecondary)
                                    .padding(.horizontal, 8)
                                    .padding(.vertical, 4)
                                    .background(Color.brandSecondary.opacity(0.15))
                                    .cornerRadius(6)
                            }
                        }
                        .cardStyle()
                    }

                    // Active subscription
                    VStack(alignment: .leading, spacing: 8) {
                        Text(NSLocalizedString("subscription_current", comment: ""))
                            .sectionTitle()

                        if let sub = viewModel.activeSubscription {
                            VStack(alignment: .leading, spacing: 12) {
                                HStack(spacing: 6) {
                                    Circle().fill(Color.success).frame(width: 8, height: 8)
                                    Text(NSLocalizedString("subscription_status", comment: "")).font(.custom("Inter-Regular", size: 14)).foregroundColor(.textSecondary)
                                    Spacer()
                                    Text(NSLocalizedString("subscription_active", comment: "")).font(.custom("Inter-Semibold", size: 14)).foregroundColor(.success)
                                }
                                Divider().background(Color.divider)
                                labeledRow(NSLocalizedString("subscription_start_time", comment: ""), sub.startTime)
                                Divider().background(Color.divider)
                                labeledRow(NSLocalizedString("subscription_end_time", comment: ""), sub.endTime)
                                Divider().background(Color.divider)
                                labeledRow(NSLocalizedString("subscription_price", comment: ""), "\u{00A5}\(String(format: "%.2f", sub.price))")
                                if let method = sub.paymentMethod {
                                    Divider().background(Color.divider)
                                    labeledRow(NSLocalizedString("subscription_payment_method", comment: ""), method)
                                }
                            }
                            .cardStyle()
                            .overlay(
                                RoundedRectangle(cornerRadius: 16)
                                    .fill(Color.brandPrimary.opacity(0))
                                    .overlay(
                                        Rectangle().fill(Color.brandPrimary).frame(width: 3),
                                        alignment: .leading
                                    )
                                    .clipShape(RoundedRectangle(cornerRadius: 16))
                            )
                        } else {
                            Text(NSLocalizedString("subscription_no_active", comment: ""))
                                .font(.custom("Inter-Regular", size: 14))
                                .foregroundColor(.textSecondary)
                                .cardStyle()
                                .frame(maxWidth: .infinity)
                        }
                    }

                    // History
                    VStack(alignment: .leading, spacing: 0) {
                        Text(NSLocalizedString("subscription_history", comment: ""))
                            .sectionTitle()
                            .padding(.bottom, 8)

                        VStack(spacing: 0) {
                            ForEach(viewModel.subscriptions.filter { $0.status != "active" }) { sub in
                                HStack {
                                    VStack(alignment: .leading, spacing: 4) {
                                        Text("\(sub.startTime) - \(sub.endTime)")
                                            .font(.custom("Inter-Regular", size: 13))
                                            .foregroundColor(.textSecondary)
                                    }
                                    Spacer()
                                    Text("\u{00A5}\(String(format: "%.2f", sub.price))")
                                        .font(.custom("Inter-Semibold", size: 14))
                                        .foregroundColor(.textPrimary)
                                }
                                .padding(.horizontal, 16)
                                .frame(height: 48)
                                if sub.id != viewModel.subscriptions.last?.id {
                                    Divider().background(Color.divider).padding(.leading, 16)
                                }
                            }
                        }
                        .background(Color.bgCard)
                        .cornerRadius(16)
                    }
                }
                .padding(16)
            }
        }
        .navigationTitle(NSLocalizedString("subscription_title", comment: ""))
        .toolbarBackground(Color.bgPrimary.opacity(0.9), for: .navigationBar)
        .refreshable { await viewModel.load() }
        .task { await viewModel.load() }
    }

    private func labeledRow(_ label: String, _ value: String) -> some View {
        HStack {
            Text(label).font(.custom("Inter-Regular", size: 14)).foregroundColor(.textSecondary)
            Spacer()
            Text(value).font(.custom("Inter-Regular", size: 14)).foregroundColor(.textPrimary)
        }
    }
}
