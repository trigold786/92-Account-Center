import SwiftUI

@MainActor
class LoginViewModel: ObservableObject {
    @Published var phoneNumber = ""
    @Published var password = ""
    @Published var verificationCode = ""
    @Published var loginMode: LoginMode = .password
    @Published var isLoading = false
    @Published var countdownSeconds = 0
    @Published var errorMessage: String?
    
    enum LoginMode {
        case password
        case verificationCode
    }
    
    private var countdownTimer: Timer?
    private let authManager: AuthManager
    
    init(authManager: AuthManager = .shared) {
        self.authManager = authManager
    }
    
    func login() async {
        guard !isLoading else { return }
        isLoading = true
        errorMessage = nil
        
        do {
            let request: LoginRequest
            switch loginMode {
            case .password:
                request = LoginRequest(phoneNumber: phoneNumber, password: password)
            case .verificationCode:
                request = LoginRequest(phoneNumber: phoneNumber, code: verificationCode)
            }
            
            let token = try await APIClient.shared.login(request: request)
            authManager.login(token: token)
        } catch let error as APIError {
            errorMessage = error.errorDescription
        } catch {
            errorMessage = "登录失败，请稍后重试"
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
