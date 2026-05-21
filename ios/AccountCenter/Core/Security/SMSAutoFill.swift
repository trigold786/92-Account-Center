import UIKit

class SMSAutoFill {
    static let shared = SMSAutoFill()

    var oneTimeCodeContentType: UITextContentType {
        return .oneTimeCode
    }

    func configureTextFieldForOTP(_ textField: UITextField) {
        textField.textContentType = .oneTimeCode
        textField.keyboardType = .numberPad
    }

    func generateVerificationHint(phoneNumber: String) -> String {
        let last4 = String(phoneNumber.suffix(4))
        let maskedNumber = "***-***-\(last4)"
        return "Verification code sent to \(maskedNumber)"
    }

    func configureTextFieldGroup(_ textFields: [UITextField]) {
        for textField in textFields {
            textField.textContentType = .oneTimeCode
            textField.keyboardType = .numberPad
        }
    }
}
