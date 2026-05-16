import SwiftUI

@MainActor
class RegisterViewModel: ObservableObject {
    @Published var phoneNumber = ""
    @Published var verificationCode = ""
    @Published var password = ""
    @Published var confirmPassword = ""
    @Published var accountId = ""
    @Published var agreeToTerms = false
    @Published var referralCode = ""
    
    @Published var isLoading = false
    @Published var countdownSeconds = 0
    @Published var errorMessage: String?
    
    private var countdownTimer: Timer?
    @EnvironmentObject var authManager: AuthManager
    
    func register() async {
        guard !isLoading, validateInputs() else { return }
        isLoading = true
        errorMessage = nil
        
        do {
            let request = RegisterRequest(
                phoneNumber: phoneNumber,
                accountId: accountId,
                password: password,
                agreeToTerms: agreeToTerms,
                referralCode: referralCode.isEmpty ? nil : referralCode
            )
            let _ = try await APIClient.shared.register(request: request)
            
            let loginReq = LoginRequest(phoneNumber: phoneNumber, password: password)
            let token = try await APIClient.shared.login(request: loginReq)
            authManager.login(token: token)
        } catch let error as APIError {
            errorMessage = error.errorDescription
        } catch {
            errorMessage = "注册失败，请稍后重试"
        }
        
        isLoading = false
    }
    
    func sendVerificationCode() async {
        guard !isLoading, countdownSeconds == 0 else { return }
        isLoading = true
        errorMessage = nil
        
        do {
            let request = SMSSendRequest(phoneNumber: phoneNumber)
            try await APIClient.shared.sendSMS(request: request)
            startCountdown()
        } catch let error as APIError {
            errorMessage = error.errorDescription
        } catch {
            errorMessage = "发送失败，请稍后重试"
        }
        
        isLoading = false
    }
    
    private func validateInputs() -> Bool {
        guard phoneNumber.count == 11 else {
            errorMessage = "请输入正确的手机号"
            return false
        }
        guard password.count >= 8 else {
            errorMessage = "密码至少 8 位"
            return false
        }
        guard password == confirmPassword else {
            errorMessage = "两次密码不一致"
            return false
        }
        guard !accountId.isEmpty else {
            errorMessage = "请输入账户ID"
            return false
        }
        guard agreeToTerms else {
            errorMessage = "请同意服务条款"
            return false
        }
        return true
    }
    
    private func startCountdown() {
        countdownSeconds = 60
        countdownTimer?.invalidate()
        
        countdownTimer = Timer.scheduledTimer(withTimeInterval: 1, repeats: true) { [weak self] _ in
            guard let self = self else { return }
            DispatchQueue.main.async {
                if self.countdownSeconds > 0 {
                    self.countdownSeconds -= 1
                } else {
                    self.countdownTimer?.invalidate()
                }
            }
        }
    }
}
