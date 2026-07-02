import SwiftUI

struct CreditsView: View {
    @StateObject private var viewModel = CreditsViewModel()

    var body: some View {
        ZStack {
            Color.bgPrimary.ignoresSafeArea()

            ScrollView {
                VStack(spacing: 12) {
                    if viewModel.isLoading {
                        ProgressView().padding()
                    }

                    if let errorMessage = viewModel.errorMessage {
                        Text(errorMessage).font(.custom("Inter-Regular", size: 12)).foregroundColor(.danger)
                    }

                    if let account = viewModel.account {
                        VStack(spacing: 4) {
                            Text(NSLocalizedString("credits_current", comment: ""))
                                .font(.custom("Inter-Regular", size: 13))
                                .foregroundColor(.textSecondary)
                            Text("\u{00A5}\(String(format: "%.2f", account.balance))")
                                .font(.custom("SpaceGrotesk-Bold", size: 36))
                                .foregroundStyle(Color.brandGradient)
                            Text(account.status == "active" ? NSLocalizedString("credits_account_active", comment: "") : account.status)
                                .font(.custom("Inter-Regular", size: 12))
                                .foregroundColor(.success)
                        }
                        .cardStyle()
                        .frame(maxWidth: .infinity)
                    }

                    // MARK: - Quick Actions
                    VStack(alignment: .leading, spacing: 0) {
                        Text(NSLocalizedString("credits_quick_actions", comment: "")).sectionTitle().padding(.bottom, 8)
                        Button(NSLocalizedString("credits_daily_checkin", comment: "")) {
                            Task { await viewModel.earnCredits(amount: 1, reason: "daily_checkin") }
                        }
                        .buttonStyle(.borderedProminent)
                        .frame(maxWidth: .infinity)
                    }

                    if let referral = viewModel.referral {
                        VStack(alignment: .leading, spacing: 8) {
                            Text(NSLocalizedString("credits_referral", comment: "")).font(.custom("Inter-Semibold", size: 15)).foregroundColor(.textPrimary)
                            labeledRow(NSLocalizedString("credits_referral_count", comment: ""), "\(referral.totalReferees)")
                            labeledRow(NSLocalizedString("credits_active_friends", comment: ""), "\(referral.activeReferees)")
                            labeledRow(NSLocalizedString("credits_total_earned", comment: ""), "\u{00A5}\(String(format: "%.2f", referral.totalEarned))")
                            Button(action: { Task { await viewModel.generateLink() } }) {
                                Label(NSLocalizedString("credits_copy_referral", comment: ""), systemImage: "link")
                                    .font(.custom("Inter-Regular", size: 14))
                                    .foregroundColor(.brandPrimary)
                            }
                            .buttonStyle(.bordered)
                            .tint(.brandPrimary)
                        }
                        .cardStyle()
                    }

                    VStack(alignment: .leading, spacing: 0) {
                        Text(NSLocalizedString("credits_transaction_history", comment: ""))
                            .sectionTitle()
                            .padding(.bottom, 8)

                        if viewModel.transactions.isEmpty {
                            Text(NSLocalizedString("credits_no_transactions", comment: ""))
                                .font(.custom("Inter-Regular", size: 14))
                                .foregroundColor(.textSecondary)
                                .padding(24)
                                .frame(maxWidth: .infinity)
                                .background(Color.bgCard)
                                .cornerRadius(16)
                        } else {
                            VStack(spacing: 0) {
                                ForEach(viewModel.transactions) { txn in
                                    HStack(spacing: 12) {
                                        Image(systemName: viewModel.transactionTypeIcon(txn.type))
                                            .font(.system(size: 16))
                                            .foregroundColor(viewModel.transactionTypeColor(txn.type))
                                            .frame(width: 24)
                                        VStack(alignment: .leading, spacing: 2) {
                                            Text(txn.details ?? txn.type)
                                                .font(.custom("Inter-Regular", size: 14))
                                                .foregroundColor(.textPrimary)
                                            Text(txn.createdAt)
                                                .font(.custom("Inter-Regular", size: 11))
                                                .foregroundColor(.textSecondary)
                                        }
                                        Spacer()
                                        Text("\(txn.amount >= 0 ? "+" : "")\u{00A5}\(String(format: "%.2f", txn.amount))")
                                            .font(.custom("Inter-Semibold", size: 14))
                                            .foregroundColor(viewModel.transactionTypeColor(txn.type))
                                    }
                                    .padding(.horizontal, 16)
                                    .frame(minHeight: 52)
                                    if txn.id != viewModel.transactions.last?.id {
                                        Divider().background(Color.divider).padding(.leading, 52)
                                    }
                                }
                            }
                            .background(Color.bgCard)
                            .cornerRadius(16)
                        }

                        if viewModel.hasMore {
                            Button(NSLocalizedString("credits_load_more", comment: "")) { Task { await viewModel.loadMore() } }
                                .font(.custom("Inter-Regular", size: 14))
                                .foregroundColor(.brandSecondary)
                                .padding(.top, 8)
                        }
                    }
                }
                .padding(16)
            }
        }
        .navigationTitle(NSLocalizedString("credits_title", comment: ""))
        .toolbarBackground(Color.bgPrimary.opacity(0.9), for: .navigationBar)
        .refreshable { await viewModel.load() }
        .task { await viewModel.load() }
    }

    private func labeledRow(_ label: String, _ value: String) -> some View {
        HStack {
            Text(label).font(.custom("Inter-Regular", size: 14)).foregroundColor(.textSecondary)
            Spacer()
            Text(value).font(.custom("Inter-Semibold", size: 14)).foregroundColor(.textPrimary)
        }
    }
}
