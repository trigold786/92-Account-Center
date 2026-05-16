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
                        Text("创建账户")
                            .font(.custom("SpaceGrotesk-Bold", size: 28))
                            .foregroundColor(.textPrimary)
                        Text("注册新账户以使用完整功能")
                            .font(.custom("Inter-Regular", size: 14))
                            .foregroundColor(.textSecondary)
                    }
                    .padding(.bottom, 32)

                    TextField("手机号", text: $viewModel.phoneNumber)
                        .keyboardType(.numberPad)
                        .textContentType(.telephoneNumber)
                        .font(.custom("Inter-Regular", size: 15))
                        .glowingInput()

                    HStack(spacing: 8) {
                        TextField("验证码", text: $viewModel.verificationCode)
                            .keyboardType(.numberPad)
                            .font(.custom("Inter-Regular", size: 15))
                            .glowingInput()

                        Button(viewModel.countdownSeconds > 0 ? "\(viewModel.countdownSeconds)s" : "发送验证码") {
                            Task { await viewModel.sendVerificationCode() }
                        }
                        .font(.custom("Inter-Regular", size: 13))
                        .foregroundColor(.brandSecondary)
                        .disabled(viewModel.countdownSeconds > 0 || viewModel.phoneNumber.isEmpty)
                    }

                    TextField("账户ID", text: $viewModel.accountId)
                        .textContentType(.username)
                        .font(.custom("Inter-Regular", size: 15))
                        .glowingInput()

                    SecureField("密码", text: $viewModel.password)
                        .textContentType(.newPassword)
                        .font(.custom("Inter-Regular", size: 15))
                        .glowingInput()

                    SecureField("确认密码", text: $viewModel.confirmPassword)
                        .textContentType(.newPassword)
                        .font(.custom("Inter-Regular", size: 15))
                        .glowingInput()

                    TextField("推荐码（可选）", text: $viewModel.referralCode)
                        .font(.custom("Inter-Regular", size: 15))
                        .glowingInput()

                    Toggle(isOn: $viewModel.agreeToTerms) {
                        Text("我已阅读并同意服务条款和隐私政策")
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
                            Text("注册")
                                .font(.custom("Inter-Semibold", size: 16))
                        }
                    }
                    .gradientButton()
                    .disabled(viewModel.isLoading)

                    HStack(spacing: 4) {
                        Text("已有账号？")
                            .font(.custom("Inter-Regular", size: 13))
                            .foregroundColor(.textSecondary)
                        Button(action: { dismiss() }) {
                            Text("立即登录")
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
