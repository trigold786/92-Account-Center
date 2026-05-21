import SwiftUI

struct EmptyStateView: View {
    let icon: String
    let title: String
    let description: String
    var actionText: String? = nil
    var onAction: (() -> Void)? = nil

    var body: some View {
        VStack(spacing: 16) {
            Text(icon)
                .font(.system(size: 48))

            Text(title)
                .font(.headline)
                .foregroundColor(Color("TextPrimary"))

            Text(description)
                .font(.subheadline)
                .foregroundColor(Color("TextSecondary"))
                .multilineTextAlignment(.center)
                .padding(.horizontal, 32)

            if let actionText = actionText, let onAction = onAction {
                Button(action: onAction) {
                    Text(actionText)
                        .font(.subheadline.weight(.semibold))
                        .foregroundColor(.white)
                        .padding(.horizontal, 32)
                        .padding(.vertical, 10)
                        .background(
                            LinearGradient(
                                colors: [Color("BrandPrimary"), Color("BrandSecondary")],
                                startPoint: .leading,
                                endPoint: .trailing
                            )
                        )
                        .cornerRadius(8)
                }
                .padding(.top, 8)
            }
        }
        .padding(32)
    }
}
