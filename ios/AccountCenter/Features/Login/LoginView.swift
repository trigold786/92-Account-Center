import SwiftUI

struct LoginView: View {
    @StateObject private var viewModel = LoginViewModel()
    @Environment(\.presentationMode) var presentationMode
    
    var body: some View {
        VStack(spacing: 24) {
            Spacer()
            
            VStack(spacing: 12) {
                Text("账户中心")
                    .font(.largeTitle)
                    .fontWeight(.bold)
                
                Text("登录您的账户")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
            }
            .padding(.bottom, 48)
            
            VStack(spacing: 16) {
                TextField("手机号", text: $viewModel.phoneNumber)
                    .keyboardType(.numberPad)
                    .textContentType(.telephoneNumber)
                    .padding()
                    .background(Color(.systemGray6))
                    .cornerRadius(8)
                
                Picker("登录方式", selection: $viewModel.loginMode) {
                    Text("密码登录").tag(LoginViewModel.LoginMode.password)
                    Text("验证码登录").tag(LoginViewModel.LoginMode.verificationCode)
                }
                .pickerStyle(.segmented)
                
                if viewModel.loginMode == .password {
                    SecureField("密码", text: $viewModel.password)
                        .textContentType(.password)
                        .padding()
                        .background(Color(.systemGray6))
                        .cornerRadius(8)
                } else {
                    HStack {
                        TextField("验证码", text: $viewModel.verificationCode)
                            .keyboardType(.numberPad)
                            .padding()
                            .background(Color(.systemGray6))
                            .cornerRadius(8)
                        
                        Button(viewModel.countdownSeconds > 0 ? "\(viewModel.countdownSeconds)s" : "发送验证码") {
                            Task {
                                await viewModel.sendVerificationCode()
                            }
                        }
                        .disabled(viewModel.countdownSeconds > 0 || viewModel.phoneNumber.isEmpty)
                        .padding()
                        .background(Color.blue)
                        .foregroundColor(.white)
                        .cornerRadius(8)
                    }
                }
                
                if let errorMessage = viewModel.errorMessage {
                    Text(errorMessage)
                        .foregroundColor(.red)
                        .font(.caption)
                }
                
                Button(action: {
                    Task {
                        await viewModel.login()
                    }
                }) {
                    if viewModel.isLoading {
                        ProgressView()
                    } else {
                        Text("登录")
                            .frame(maxWidth: .infinity)
                    }
                }
                .padding()
                .background(Color.blue)
                .foregroundColor(.white)
                .cornerRadius(8)
                .disabled(viewModel.isLoading)
            }
            .padding(.horizontal, 32)
            
            HStack {
                Text("还没有账号？")
                    .foregroundStyle(.secondary)
                NavigationLink(destination: RegisterView()) {
                    Text("立即注册")
                        .fontWeight(.semibold)
                }
            }
            .padding(.top, 24)
            
            Spacer()
        }
        .padding()
    }
}

struct LoginView_Previews: PreviewProvider {
    static var previews: some View {
        LoginView()
    }
}
