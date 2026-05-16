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
                            Text("当前积分")
                                .font(.custom("Inter-Regular", size: 13))
                                .foregroundColor(.textSecondary)
                            Text("\u{00A5}\(String(format: "%.2f", account.balance))")
                                .font(.custom("SpaceGrotesk-Bold", size: 36))
                                .foregroundStyle(Color.brandGradient)
                            Text(account.status == "active" ? "账户正常" : account.status)
                                .font(.custom("Inter-Regular", size: 12))
                                .foregroundColor(.success)
                        }
                        .cardStyle()
                        .frame(maxWidth: .infinity)
                    }

                    if let referral = viewModel.referral {
                        VStack(alignment: .leading, spacing: 8) {
                            Text("邀请推广").font(.custom("Inter-Semibold", size: 15)).foregroundColor(.textPrimary)
                            labeledRow("邀请人数", "\(referral.totalReferees)")
                            labeledRow("活跃好友", "\(referral.activeReferees)")
                            labeledRow("累计收益", "\u{00A5}\(String(format: "%.2f", referral.totalEarned))")
                            Button(action: { Task { await viewModel.generateLink() } }) {
                                Label("复制推荐链接", systemImage: "link")
                                    .font(.custom("Inter-Regular", size: 14))
                                    .foregroundColor(.brandPrimary)
                            }
                            .buttonStyle(.bordered)
                            .tint(.brandPrimary)
                        }
                        .cardStyle()
                    }

                    VStack(alignment: .leading, spacing: 0) {
                        Text("交易记录")
                            .sectionTitle()
                            .padding(.bottom, 8)

                        if viewModel.transactions.isEmpty {
                            Text("暂无交易记录")
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
                            Button("加载更多...") { Task { await viewModel.loadMore() } }
                                .font(.custom("Inter-Regular", size: 14))
                                .foregroundColor(.brandSecondary)
                                .padding(.top, 8)
                        }
                    }
                }
                .padding(16)
            }
        }
        .navigationTitle("积分中心")
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
