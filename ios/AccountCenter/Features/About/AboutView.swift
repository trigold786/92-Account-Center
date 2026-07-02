import SwiftUI

struct AboutView: View {
    var body: some View {
        ZStack {
            Color.bgPrimary.ignoresSafeArea()

            VStack(spacing: 0) {
                Spacer().frame(height: 48)

                VStack(spacing: 12) {
                    Image(systemName: "person.circle.fill")
                        .font(.system(size: 56))
                        .foregroundStyle(Color.brandGradient)
                    Text(NSLocalizedString("about_account_center", comment: ""))
                        .font(.custom("SpaceGrotesk-Bold", size: 22))
                        .foregroundColor(.textPrimary)
                    HStack(spacing: 4) {
                        Text("Version")
                            .font(.custom("Inter-Regular", size: 13))
                            .foregroundColor(.textSecondary)
                        Text(Bundle.main.object(forInfoDictionaryKey: "CFBundleShortVersionString") as? String ?? "1.0")
                            .font(.custom("Inter-Semibold", size: 13))
                            .foregroundColor(.textPrimary)
                        Text("Build")
                            .font(.custom("Inter-Regular", size: 13))
                            .foregroundColor(.textSecondary)
                            .padding(.leading, 8)
                        Text(Bundle.main.object(forInfoDictionaryKey: "CFBundleVersion") as? String ?? "1")
                            .font(.custom("Inter-Semibold", size: 13))
                            .foregroundColor(.textPrimary)
                    }
                }
                .padding(.bottom, 48)

                VStack(spacing: 0) {
                    aboutRow(NSLocalizedString("about_version", comment: ""), Bundle.main.object(forInfoDictionaryKey: "CFBundleShortVersionString") as? String ?? "1.0")
                    Divider().background(Color.divider).padding(.leading, 16)
                    aboutRow(NSLocalizedString("about_build", comment: ""), Bundle.main.object(forInfoDictionaryKey: "CFBundleVersion") as? String ?? "1")
                }
                .background(Color.bgCard).cornerRadius(16)
                .padding(.horizontal, 16)

                VStack(spacing: 0) {
                    legalRow(NSLocalizedString("about_terms", comment: ""))
                    Divider().background(Color.divider).padding(.leading, 16)
                    legalRow(NSLocalizedString("about_privacy", comment: ""))
                }
                .background(Color.bgCard).cornerRadius(16)
                .padding(.horizontal, 16)
                .padding(.top, 12)

                Spacer()

                Text("\u{00A9} 2026 Account Center. All rights reserved.")
                    .font(.custom("Inter-Regular", size: 11))
                    .foregroundColor(.textSecondary)
                    .padding(.bottom, 24)
            }
        }
        .navigationTitle(NSLocalizedString("about_title", comment: ""))
        .toolbarBackground(Color.bgPrimary.opacity(0.9), for: .navigationBar)
    }

    private func aboutRow(_ label: String, _ value: String) -> some View {
        HStack {
            Text(label).font(.custom("Inter-Regular", size: 14)).foregroundColor(.textSecondary)
            Spacer()
            Text(value).font(.custom("Inter-Semibold", size: 14)).foregroundColor(.textPrimary)
        }
        .padding(.horizontal, 16).frame(height: 48)
    }

    private func legalRow(_ label: String) -> some View {
        HStack {
            Text(label).font(.custom("Inter-Regular", size: 14)).foregroundColor(.brandSecondary)
            Spacer()
            Image(systemName: "chevron.right").font(.system(size: 12)).foregroundColor(.textSecondary)
        }
        .padding(.horizontal, 16).frame(height: 48)
    }
}
