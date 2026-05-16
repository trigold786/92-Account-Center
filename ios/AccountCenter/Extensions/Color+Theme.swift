import SwiftUI

extension Color {
    static let bgPrimary = Color(hex: "#0D1117")
    static let bgCard = Color(hex: "#161B22")
    static let bgInput = Color(hex: "#161B22")
    static let brandPrimary = Color(hex: "#6C63FF")
    static let brandSecondary = Color(hex: "#00D4FF")
    static let textPrimary = Color(hex: "#DEFFFFFF")
    static let textSecondary = Color(hex: "#8B949E")
    static let divider = Color(hex: "#21262D")
    static let danger = Color(hex: "#FF4757")
    static let success = Color(hex: "#2ED573")
    static let tierFree = Color(hex: "#8B949E")
    static let tierBasic = Color(hex: "#6C63FF")
    static let tierPremium = Color(hex: "#FF9800")
    static let tierEnterprise = Color(hex: "#7B1FA2")

    static let brandGradient = LinearGradient(
        colors: [.brandPrimary, .brandSecondary],
        startPoint: .topLeading,
        endPoint: .bottomTrailing
    )

    init(hex: String) {
        let hex = hex.trimmingCharacters(in: CharacterSet.alphanumerics.inverted)
        var int: UInt64 = 0
        Scanner(string: hex).scanHexInt64(&int)
        let a, r, g, b: UInt64
        switch hex.count {
        case 8:
            (a, r, g, b) = ((int >> 24) & 0xFF, (int >> 16) & 0xFF, (int >> 8) & 0xFF, int & 0xFF)
        default:
            (a, r, g, b) = (0xFF as UInt64, (int >> 16) & 0xFF, (int >> 8) & 0xFF, int & 0xFF)
        }
        self.init(
            .sRGB,
            red: Double(r) / 255,
            green: Double(g) / 255,
            blue: Double(b) / 255,
            opacity: Double(a) / 255
        )
    }
}
