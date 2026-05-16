import SwiftUI

struct SecurityView: View {
    @StateObject private var viewModel = SecurityViewModel()

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

                    VStack(alignment: .leading, spacing: 0) {
                        Text("风险事件").sectionTitle().padding(.bottom, 8)

                        if viewModel.riskEvents.isEmpty {
                            Text("暂无风险事件")
                                .font(.custom("Inter-Regular", size: 14)).foregroundColor(.textSecondary)
                                .padding(24).frame(maxWidth: .infinity).background(Color.bgCard).cornerRadius(16)
                        } else {
                            VStack(spacing: 0) {
                                ForEach(viewModel.riskEvents) { event in
                                    HStack(spacing: 12) {
                                        Circle().fill(viewModel.riskLevelColor(event.riskLevel)).frame(width: 10, height: 10)
                                        VStack(alignment: .leading, spacing: 2) {
                                            Text(event.eventType).font(.custom("Inter-Regular", size: 14)).foregroundColor(.textPrimary)
                                            Text(event.createdAt).font(.custom("Inter-Regular", size: 11)).foregroundColor(.textSecondary)
                                        }
                                        Spacer()
                                        Text(event.riskLevel.uppercased())
                                            .font(.custom("Inter-Semibold", size: 10))
                                            .foregroundColor(viewModel.riskLevelColor(event.riskLevel))
                                            .padding(.horizontal, 6).padding(.vertical, 3)
                                            .background(viewModel.riskLevelColor(event.riskLevel).opacity(0.15))
                                            .cornerRadius(4)
                                    }
                                    .padding(.horizontal, 16).frame(minHeight: 52)
                                    if event.id != viewModel.riskEvents.last?.id {
                                        Divider().background(Color.divider).padding(.leading, 38)
                                    }
                                }
                            }
                            .background(Color.bgCard).cornerRadius(16)
                        }
                    }

                    VStack(alignment: .leading, spacing: 0) {
                        Text("登录设备").sectionTitle().padding(.bottom, 8)

                        if viewModel.devices.isEmpty {
                            Text("暂无设备记录")
                                .font(.custom("Inter-Regular", size: 14)).foregroundColor(.textSecondary)
                                .padding(24).frame(maxWidth: .infinity).background(Color.bgCard).cornerRadius(16)
                        } else {
                            VStack(spacing: 0) {
                                ForEach(viewModel.devices) { device in
                                    HStack(spacing: 12) {
                                        Image(systemName: device.platform == "ios" ? "iphone" : "desktopcomputer")
                                            .font(.system(size: 18)).foregroundColor(.brandPrimary).frame(width: 24)
                                        VStack(alignment: .leading, spacing: 2) {
                                            Text(device.deviceName ?? device.platform)
                                                .font(.custom("Inter-Regular", size: 14)).foregroundColor(.textPrimary)
                                            if let lastActive = device.lastActiveAt {
                                                Text("最近活跃: \(lastActive)")
                                                    .font(.custom("Inter-Regular", size: 11)).foregroundColor(.textSecondary)
                                            }
                                        }
                                        Spacer()
                                        Text(device.isActive ? "活跃中" : "离线")
                                            .font(.custom("Inter-Semibold", size: 11))
                                            .foregroundColor(device.isActive ? .success : .textSecondary)
                                    }
                                    .padding(.horizontal, 16).frame(minHeight: 52)
                                    if device.id != viewModel.devices.last?.id {
                                        Divider().background(Color.divider).padding(.leading, 52)
                                    }
                                }
                            }
                            .background(Color.bgCard).cornerRadius(16)
                        }
                    }
                }
                .padding(16)
            }
        }
        .navigationTitle("安全设置")
        .toolbarBackground(Color.bgPrimary.opacity(0.9), for: .navigationBar)
        .refreshable { await viewModel.load() }
        .task { await viewModel.load() }
    }
}
