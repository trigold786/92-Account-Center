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
            .background(
                LinearGradient(
                    colors: [.brandPrimary, .brandSecondary],
                    startPoint: .topLeading,
                    endPoint: .bottomTrailing
                )
            )
            .cornerRadius(12)
    }

    func glowingInput() -> some View {
        self
            .padding()
            .background(Color.bgInput)
            .cornerRadius(12)
            .overlay(
                RoundedRectangle(cornerRadius: 12)
                    .stroke(Color.brandSecondary.opacity(0.0), lineWidth: 1)
            )
    }

    func sectionTitle() -> some View {
        self
            .font(.system(size: 13, weight: .semibold))
            .foregroundColor(.textSecondary)
            .textCase(.uppercase)
    }
}
