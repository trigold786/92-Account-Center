import SwiftUI

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
                    Text("账户中心")
                        .font(.custom("SpaceGrotesk-Bold", size: 28))
                        .foregroundColor(.textPrimary)
                    Text("登录您的账户")
                        .font(.custom("Inter-Regular", size: 14))
                        .foregroundColor(.textSecondary)
                }
                .padding(.bottom, 48)

                VStack(spacing: 16) {
                    TextField("手机号", text: $viewModel.phoneNumber)
                        .keyboardType(.numberPad)
                        .textContentType(.telephoneNumber)
                        .font(.custom("Inter-Regular", size: 15))
                        .glowingInput()

                    Picker("登录方式", selection: $viewModel.loginMode) {
                        Text("密码登录").tag(LoginViewModel.LoginMode.password)
                        Text("验证码登录").tag(LoginViewModel.LoginMode.verificationCode)
                    }
                    .pickerStyle(.segmented)

                    if viewModel.loginMode == .password {
                        SecureField("密码", text: $viewModel.password)
                            .textContentType(.password)
                            .font(.custom("Inter-Regular", size: 15))
                            .glowingInput()
                    } else {
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
                            Text("登录")
                                .font(.custom("Inter-Semibold", size: 16))
                        }
                    }
                    .gradientButton()
                    .disabled(viewModel.isLoading)
                }
                .padding(.horizontal, 32)

                HStack(spacing: 4) {
                    Text("还没有账号？")
                        .font(.custom("Inter-Regular", size: 13))
                        .foregroundColor(.textSecondary)
                    NavigationLink(destination: RegisterView()) {
                        Text("立即注册")
                            .font(.custom("Inter-Semibold", size: 13))
                            .foregroundColor(.brandSecondary)
                    }
                }
                .padding(.top, 24)

                Spacer()
            }
        }
        .navigationBarHidden(true)
    }
}
