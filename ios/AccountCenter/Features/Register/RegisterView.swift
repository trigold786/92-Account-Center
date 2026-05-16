import SwiftUI

struct RegisterView: View {
    @StateObject private var viewModel = RegisterViewModel()
    @Environment(\.presentationMode) var presentationMode
    
    var body: some View {
        ScrollView {
            VStack(spacing: 16) {
                Spacer().frame(height: 24)
                
                VStack(spacing: 8) {
                    Text("创建账户")
                        .font(.largeTitle)
                        .fontWeight(.bold)
                    Text("注册新账户以使用完整功能")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                }
                .padding(.bottom, 32)
                
                TextField("手机号", text: $viewModel.phoneNumber)
                    .keyboardType(.numberPad)
                    .textContentType(.telephoneNumber)
                    .padding()
                    .background(Color(.systemGray6))
                    .cornerRadius(8)
                
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
                
                TextField("账户ID", text: $viewModel.accountId)
                    .textContentType(.username)
                    .padding()
                    .background(Color(.systemGray6))
                    .cornerRadius(8)
                
                SecureField("密码", text: $viewModel.password)
                    .textContentType(.newPassword)
                    .padding()
                    .background(Color(.systemGray6))
                    .cornerRadius(8)
                
                SecureField("确认密码", text: $viewModel.confirmPassword)
                    .textContentType(.newPassword)
                    .padding()
                    .background(Color(.systemGray6))
                    .cornerRadius(8)
                
                TextField("推荐码（可选）", text: $viewModel.referralCode)
                    .padding()
                    .background(Color(.systemGray6))
                    .cornerRadius(8)
                
                Toggle(isOn: $viewModel.agreeToTerms) {
                    Text("我已阅读并同意服务条款和隐私政策")
                        .font(.caption)
                }
                .padding(.vertical, 8)
                
                if let errorMessage = viewModel.errorMessage {
                    Text(errorMessage)
                        .foregroundColor(.red)
                        .font(.caption)
                }
                
                Button(action: {
                    Task {
                        await viewModel.register()
                    }
                }) {
                    if viewModel.isLoading {
                        ProgressView()
                    } else {
                        Text("注册")
                            .frame(maxWidth: .infinity)
                    }
                }
                .padding()
                .background(Color.blue)
                .foregroundColor(.white)
                .cornerRadius(8)
                .disabled(viewModel.isLoading)
                
                HStack {
                    Text("已有账号？")
                        .foregroundStyle(.secondary)
                    Button(action: {
                        presentationMode.wrappedValue.dismiss()
                    }) {
                        Text("立即登录")
                            .fontWeight(.semibold)
                    }
                }
                .padding(.top, 16)
            }
            .padding(.horizontal, 32)
        }
        .navigationBarTitleDisplayMode(.inline)
        .navigationBarBackButtonHidden(false)
    }
}

struct RegisterView_Previews: PreviewProvider {
    static var previews: some View {
        RegisterView()
    }
}
