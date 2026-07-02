import SwiftUI
import UIKit

struct ScreenshotPreventer: UIViewRepresentable {
    func makeUIView(context: Context) -> UIView {
        let view = UIView()
        DispatchQueue.main.async {
            let textField = UITextField()
            textField.isSecureTextEntry = true
            textField.frame = CGRect(x: 0, y: 0, width: 1, height: 1)
            view.addSubview(textField)
        }
        return view
    }

    func updateUIView(_ uiView: UIView, context: Context) {}
}

struct PreventScreenshot: ViewModifier {
    func body(content: Content) -> some View {
        content
            .background(ScreenshotPreventer().frame(width: 0, height: 0))
    }
}

extension View {
    func preventScreenshot() -> some View {
        modifier(PreventScreenshot())
    }
}
