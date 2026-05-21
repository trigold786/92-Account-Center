import SwiftUI
import UIKit

enum DynamicTypeHelper {
    static func scaledFont(baseSize: CGFloat, weight: UIFont.Weight = .regular) -> UIFont {
        let font = UIFont.systemFont(ofSize: baseSize, weight: weight)
        let metrics: UIFontMetrics
        if #available(iOS 17, *) {
            metrics = UIFontMetrics(forTextStyle: .body)
        } else {
            metrics = UIFontMetrics.default
        }
        return metrics.scaledFont(for: font)
    }

    static func scaledSize(_ base: CGFloat) -> CGFloat {
        UIFontMetrics.default.scaledValue(for: base)
    }
}

extension View {
    func accessibilityLabel(_ key: AppStrings.Common.Type) -> some View {
        self.accessibilityLabel(Text(String(describing: key)))
    }

    func voiceOverHint(_ hint: String) -> some View {
        self.accessibilityHint(hint)
    }

    func accessableRow(label: String, value: String) -> some View {
        self.accessibilityElement(children: .combine)
            .accessibilityLabel("\(label): \(value)")
    }
}

struct ContrastChecker {
    static func luminance(red: CGFloat, green: CGFloat, blue: CGFloat) -> CGFloat {
        let r = red <= 0.03928 ? red / 12.92 : pow((red + 0.055) / 1.055, 2.4)
        let g = green <= 0.03928 ? green / 12.92 : pow((green + 0.055) / 1.055, 2.4)
        let b = blue <= 0.03928 ? blue / 12.92 : pow((blue + 0.055) / 1.055, 2.4)
        return 0.2126 * r + 0.7152 * g + 0.0722 * b
    }

    static func contrastRatio(fg: UIColor, bg: UIColor) -> CGFloat {
        var fgR: CGFloat = 0, fgG: CGFloat = 0, fgB: CGFloat = 0, fgA: CGFloat = 0
        var bgR: CGFloat = 0, bgG: CGFloat = 0, bgB: CGFloat = 0, bgA: CGFloat = 0
        fg.getRed(&fgR, green: &fgG, blue: &fgB, alpha: &fgA)
        bg.getRed(&bgR, green: &bgG, blue: &bgB, alpha: &bgA)

        let l1 = luminance(red: fgR, green: fgG, blue: fgB)
        let l2 = luminance(red: bgR, green: bgG, blue: bgB)
        let lighter = max(l1, l2)
        let darker = min(l1, l2)
        return (lighter + 0.05) / (darker + 0.05)
    }

    static func passesWCAGAA(fg: UIColor, bg: UIColor) -> Bool {
        return contrastRatio(fg: fg, bg: bg) >= 4.5
    }
}

struct MotionPreference {
    static var prefersReducedMotion: Bool {
        if #available(iOS 17, *) {
            return UIAccessibility.isReduceMotionEnabled
        } else {
            return UIAccessibility.isReduceMotionEnabled
        }
    }
}
