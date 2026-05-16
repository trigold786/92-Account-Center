import SwiftUI

@MainActor
class SecurityViewModel: ObservableObject {
    @Published var riskEvents: [RiskEvent] = []
    @Published var devices: [PushDevice] = []
    @Published var isLoading = false
    @Published var errorMessage: String?

    private let authManager: AuthManager

    init(authManager: AuthManager = .shared) {
        self.authManager = authManager
    }

    var userId: Int64? { authManager.userId }

    func load() async {
        guard let userId = userId else { return }
        isLoading = true
        errorMessage = nil

        do {
            async let events = APIClient.shared.getRiskHistory(userId: userId)
            async let devs = APIClient.shared.getUserDevices(userId: userId)
            let (historyData, deviceData) = try await (events, devs)
            riskEvents = historyData.events
            devices = deviceData.devices
        } catch let error as APIError {
            errorMessage = error.errorDescription
        } catch {
            errorMessage = "加载失败"
        }

        isLoading = false
    }

    func riskLevelColor(_ level: String) -> Color {
        switch level {
        case "critical": return .red
        case "high": return .orange
        case "medium": return .yellow
        default: return .green
        }
    }
}
