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
                        Text(NSLocalizedString("risk_events", comment: "")).sectionTitle().padding(.bottom, 8)

                        if viewModel.riskEvents.isEmpty {
                            Text(NSLocalizedString("no_data", comment: ""))
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
                        Text(NSLocalizedString("login_devices", comment: "")).sectionTitle().padding(.bottom, 8)

                        if viewModel.devices.isEmpty {
                            Text(NSLocalizedString("no_data", comment: ""))
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
                                                Text("\(NSLocalizedString("last_active", comment: "")): \(lastActive)")
                                                    .font(.custom("Inter-Regular", size: 11)).foregroundColor(.textSecondary)
                                            }
                                        }
                                        Spacer()
                                        Text(device.isActive ? NSLocalizedString("device_active", comment: "") : NSLocalizedString("device_offline", comment: ""))
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

            VStack(alignment: .leading, spacing: 0) {
                Text(NSLocalizedString("change_password", comment: "")).sectionTitle().padding(.bottom, 8)

                VStack(spacing: 12) {
                    SecureField(NSLocalizedString("current_password", comment: ""), text: $viewModel.currentPassword)
                        .textFieldStyle(.roundedBorder)
                    SecureField(NSLocalizedString("new_password", comment: ""), text: $viewModel.newPassword)
                        .textFieldStyle(.roundedBorder)
                    SecureField(NSLocalizedString("confirm_password", comment: ""), text: $viewModel.confirmPassword)
                        .textFieldStyle(.roundedBorder)
                    Button(NSLocalizedString("change_password", comment: "")) {
                        viewModel.changePassword()
                    }
                    .buttonStyle(.borderedProminent)
                    .disabled(viewModel.currentPassword.isEmpty || viewModel.newPassword.isEmpty || viewModel.newPassword != viewModel.confirmPassword)
                    if let msg = viewModel.passwordChangeMessage {
                        Text(msg)
                            .foregroundColor(viewModel.passwordChangeSuccess ? .green : .red)
                            .font(.caption)
                    }
                }
                .padding(16)
                .background(Color.bgCard)
                .cornerRadius(16)
            }
        }
        .navigationTitle(NSLocalizedString("feature_security", comment: ""))
        .toolbarBackground(Color.bgPrimary.opacity(0.9), for: .navigationBar)
        .refreshable { await viewModel.load() }
        .task { await viewModel.load() }
        .preventScreenshot()
    }
}
