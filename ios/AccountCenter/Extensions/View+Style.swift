import SwiftUI

extension View {
    func cardStyle() -> some View {
        self
            .padding(16)
            .background(Color.bgCard)
            .cornerRadius(16)
    }

    func gradientButton() -> some View {
        self
            .font(.system(size: 16, weight: .semibold))
            .foregroundColor(.white)
            .frame(maxWidth: .infinity)
            .frame(height: 52)
            .background(Color.brandGradient)
            .cornerRadius(12)
    }

    func glowingInput() -> some View {
        self
            .padding()
            .background(Color.bgInput)
            .cornerRadius(12)
    }

    func sectionTitle() -> some View {
        self
            .font(.system(size: 13, weight: .semibold))
            .foregroundColor(.textSecondary)
            .textCase(.uppercase)
    }
}
