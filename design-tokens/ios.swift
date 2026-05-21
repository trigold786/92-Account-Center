import SwiftUI

extension Color {
    static let bgPrimary = Color(red: 0.051, green: 0.067, blue: 0.09)
    static let bgSecondary = Color(red: 0.086, green: 0.106, blue: 0.133)
    static let bgCard = Color(red: 0.11, green: 0.137, blue: 0.2)
    static let bgInput = Color(red: 0.086, green: 0.106, blue: 0.133)

    static let brandPrimary = Color(red: 0.424, green: 0.388, blue: 1.0)
    static let brandSecondary = Color(red: 0.0, green: 0.831, blue: 1.0)

    static let textPrimary = Color(red: 0.902, green: 0.929, blue: 0.953)
    static let textSecondary = Color(red: 0.545, green: 0.58, blue: 0.62)
    static let textDisabled = Color(red: 0.282, green: 0.31, blue: 0.345)

    static let borderDefault = Color(red: 0.188, green: 0.212, blue: 0.239)
    static let borderMuted = Color(red: 0.129, green: 0.15, blue: 0.176)

    static let colorSuccess = Color(red: 0.18, green: 0.835, blue: 0.451)
    static let colorWarning = Color(red: 1.0, green: 0.647, blue: 0.008)
    static let colorDanger = Color(red: 1.0, green: 0.278, blue: 0.341)
    static let colorInfo = Color(red: 0.0, green: 0.831, blue: 1.0)

    static let tierFree = Color(red: 0.545, green: 0.58, blue: 0.62)
    static let tierBasic = Color(red: 0.424, green: 0.388, blue: 1.0)
    static let tierPremium = Color(red: 1.0, green: 0.596, blue: 0.0)
    static let tierEnterprise = Color(red: 0.482, green: 0.125, blue: 0.635)
}

extension Font {
    static let textStyleXS = Font.system(size: 12)
    static let textStyleSM = Font.system(size: 14)
    static let textStyleMD = Font.system(size: 16)
    static let textStyleLG = Font.system(size: 18)
    static let textStyleXL = Font.system(size: 20)
    static let textStyle2XL = Font.system(size: 24)
    static let textStyle3XL = Font.system(size: 32)
    static let textStyle4XL = Font.system(size: 48)
}

extension CGFloat {
    static let spacingXS: CGFloat = 4
    static let spacingSM: CGFloat = 8
    static let spacingMD: CGFloat = 16
    static let spacingLG: CGFloat = 24
    static let spacingXL: CGFloat = 32
    static let spacing2XL: CGFloat = 48
    static let spacing3XL: CGFloat = 64
}

extension CGFloat {
    static let radiusSM: CGFloat = 4
    static let radiusMD: CGFloat = 8
    static let radiusLG: CGFloat = 12
    static let radiusXL: CGFloat = 16
    static let radius2XL: CGFloat = 24
    static let radiusFull: CGFloat = 9999
}

enum DesignTokens {
    enum Shadows {
        static let shadowSM = Shadow(color: .black.opacity(0.3), radius: 2, x: 0, y: 1)
        static let shadowMD = Shadow(color: .black.opacity(0.4), radius: 4, x: 0, y: 4)
        static let shadowLG = Shadow(color: .black.opacity(0.5), radius: 8, x: 0, y: 8)
    }

    enum Motion {
        static let durationFast: Double = 0.15
        static let durationNormal: Double = 0.25
        static let durationSlow: Double = 0.35
    }
}
