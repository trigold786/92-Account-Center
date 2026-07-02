import SwiftUI

struct RegisterView: View {
    @StateObject private var viewModel = RegisterViewModel()
    @Environment(\.dismiss) private var dismiss

    var body: some View {
        ZStack {
            Color.bgPrimary.ignoresSafeArea()

            ScrollView {
                VStack(spacing: 16) {
                    Spacer().frame(height: 24)

                    VStack(spacing: 8) {
                        Text(NSLocalizedString("register_create_account", comment: ""))
                            .font(.custom("SpaceGrotesk-Bold", size: 28))
                            .foregroundColor(.textPrimary)
                        Text(NSLocalizedString("register_subtitle", comment: ""))
                            .font(.custom("Inter-Regular", size: 14))
                            .foregroundColor(.textSecondary)
                    }
                    .padding(.bottom, 32)

                    TextField(NSLocalizedString("register_phone_label", comment: ""), text: $viewModel.phoneNumber)
                        .keyboardType(.numberPad)
                        .textContentType(.telephoneNumber)
                        .font(.custom("Inter-Regular", size: 15))
                        .glowingInput()

                    HStack(spacing: 8) {
                        TextField(NSLocalizedString("register_code_label", comment: ""), text: $viewModel.verificationCode)
                            .keyboardType(.numberPad)
                            .font(.custom("Inter-Regular", size: 15))
                            .glowingInput()

                        Button(viewModel.countdownSeconds > 0 ? "\(viewModel.countdownSeconds)s" : NSLocalizedString("register_send_code", comment: "")) {
                            Task { await viewModel.sendVerificationCode() }
                        }
                        .font(.custom("Inter-Regular", size: 13))
                        .foregroundColor(.brandSecondary)
                        .disabled(viewModel.countdownSeconds > 0 || viewModel.phoneNumber.isEmpty)
                    }

                    TextField(NSLocalizedString("register_account_id", comment: ""), text: $viewModel.accountId)
                        .textContentType(.username)
                        .font(.custom("Inter-Regular", size: 15))
                        .glowingInput()

                    SecureField(NSLocalizedString("register_password_label", comment: ""), text: $viewModel.password)
                        .textContentType(.newPassword)
                        .font(.custom("Inter-Regular", size: 15))
                        .glowingInput()

                    SecureField(NSLocalizedString("register_confirm_password", comment: ""), text: $viewModel.confirmPassword)
                        .textContentType(.newPassword)
                        .font(.custom("Inter-Regular", size: 15))
                        .glowingInput()

                    TextField(NSLocalizedString("register_referral_code", comment: ""), text: $viewModel.referralCode)
                        .font(.custom("Inter-Regular", size: 15))
                        .glowingInput()

                    Toggle(isOn: $viewModel.agreeToTerms) {
                        Text(NSLocalizedString("register_agree_terms", comment: ""))
                            .font(.custom("Inter-Regular", size: 12))
                            .foregroundColor(.textSecondary)
                    }
                    .tint(.brandPrimary)

                    if let errorMessage = viewModel.errorMessage {
                        Text(errorMessage)
                            .font(.custom("Inter-Regular", size: 12))
                            .foregroundColor(.danger)
                    }

                    Button(action: { Task { await viewModel.register() } }) {
                        if viewModel.isLoading {
                            ProgressView()
                        } else {
                            Text(NSLocalizedString("register_login", comment: ""))
                                .font(.custom("Inter-Semibold", size: 16))
                        }
                    }
                    .gradientButton()
                    .disabled(viewModel.isLoading)

                    HStack(spacing: 4) {
                        Text(NSLocalizedString("register_has_account", comment: ""))
                            .font(.custom("Inter-Regular", size: 13))
                            .foregroundColor(.textSecondary)
                        Button(action: { dismiss() }) {
                            Text(NSLocalizedString("register_login_now", comment: ""))
                                .font(.custom("Inter-Semibold", size: 13))
                                .foregroundColor(.brandSecondary)
                        }
                    }
                    .padding(.top, 16)
                }
                .padding(.horizontal, 32)
            }
        }
        .navigationBarTitleDisplayMode(.inline)
    }
}
