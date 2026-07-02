import SwiftUI
import UIKit

struct LoginView: View {
    @StateObject private var viewModel = LoginViewModel()

    var body: some View {
        ZStack {
            Color.bgPrimary.ignoresSafeArea()

            VStack(spacing: 0) {
                Spacer()

                VStack(spacing: 8) {
                    Image(systemName: "person.circle.fill")
                        .font(.system(size: 48))
                        .foregroundStyle(Color.brandGradient)
                    Text(NSLocalizedString("login_title", comment: ""))
                        .font(.custom("SpaceGrotesk-Bold", size: 28))
                        .foregroundColor(.textPrimary)
                    Text(NSLocalizedString("login_subtitle", comment: ""))
                        .font(.custom("Inter-Regular", size: 14))
                        .foregroundColor(.textSecondary)
                }
                .padding(.bottom, 48)

                VStack(spacing: 16) {
                    TextField(NSLocalizedString("login_phone_label", comment: ""), text: $viewModel.phoneNumber)
                        .keyboardType(.numberPad)
                        .textContentType(.telephoneNumber)
                        .font(.custom("Inter-Regular", size: 15))
                        .glowingInput()

                    Picker(NSLocalizedString("login_password_mode", comment: ""), selection: $viewModel.loginMode) {
                        Text(NSLocalizedString("login_password_mode", comment: "")).tag(LoginViewModel.LoginMode.password)
                        Text(NSLocalizedString("login_code_mode", comment: "")).tag(LoginViewModel.LoginMode.verificationCode)
                    }
                    .pickerStyle(.segmented)

                    if viewModel.loginMode == .password {
                        SecureField(NSLocalizedString("login_password_placeholder", comment: ""), text: $viewModel.password)
                            .textContentType(.password)
                            .font(.custom("Inter-Regular", size: 15))
                            .glowingInput()
                    } else {
                        HStack(spacing: 8) {
                            TextField(NSLocalizedString("register_code_placeholder", comment: ""), text: $viewModel.verificationCode)
                                .keyboardType(.numberPad)
                                .font(.custom("Inter-Regular", size: 15))
                                .glowingInput()

                            Button(viewModel.countdownSeconds > 0 ? "\(viewModel.countdownSeconds)s" : NSLocalizedString("send_code", comment: "")) {
                                Task { await viewModel.sendVerificationCode() }
                            }
                            .font(.custom("Inter-Regular", size: 13))
                            .foregroundColor(.brandSecondary)
                            .disabled(viewModel.countdownSeconds > 0 || viewModel.phoneNumber.isEmpty)
                        }
                    }

                    if let errorMessage = viewModel.errorMessage {
                        Text(errorMessage)
                            .font(.custom("Inter-Regular", size: 12))
                            .foregroundColor(.danger)
                    }

                    Button(action: { Task { await viewModel.login() } }) {
                        if viewModel.isLoading {
                            ProgressView()
                        } else {
                            Text(NSLocalizedString("login_button", comment: ""))
                                .font(.custom("Inter-Semibold", size: 16))
                        }
                    }
                    .gradientButton()
                    .disabled(viewModel.isLoading)
                }
                .padding(.horizontal, 32)

                HStack(spacing: 4) {
                    Text(NSLocalizedString("login_no_account", comment: ""))
                        .font(.custom("Inter-Regular", size: 13))
                        .foregroundColor(.textSecondary)
                    NavigationLink(destination: RegisterView()) {
                        Text(NSLocalizedString("login_register", comment: ""))
                            .font(.custom("Inter-Semibold", size: 13))
                            .foregroundColor(.brandSecondary)
                    }
                }
                .padding(.top, 24)

                Spacer()
            }
        }
        .navigationBarHidden(true)
        .preventScreenshot()
    }
}
