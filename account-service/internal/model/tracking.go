package model

import "encoding/json"

type EventName string

const (
	EventRegisterStart   EventName = "register_start"
	EventRegisterSMSSent EventName = "register_sms_sent"
	EventRegisterSuccess EventName = "register_success"
	EventRegisterFail    EventName = "register_fail"

	EventLoginStart      EventName = "login_start"
	EventLoginSuccess    EventName = "login_success"
	EventLoginFail       EventName = "login_fail"
	EventLoginMFARequired EventName = "login_mfa_required"
	EventLoginMFASuccess EventName = "login_mfa_success"
	EventLoginMFAFail    EventName = "login_mfa_fail"

	EventPasswordChangeStart   EventName = "password_change_start"
	EventPasswordChangeSuccess EventName = "password_change_success"
	EventPasswordChangeFail    EventName = "password_change_fail"

	EventAccountDeleteApply  EventName = "account_delete_apply"
	EventAccountDeleteCancel EventName = "account_delete_cancel"
	EventAccountDeleteSuccess EventName = "account_delete_success"

	EventKYBApply        EventName = "kyb_apply"
	EventKYBVerifySuccess EventName = "kyb_verify_success"
	EventKYBVerifyFail   EventName = "kyb_verify_fail"
)

type TrackingEvent struct {
	EventName  EventName     `json:"event_name"`
	Properties TrackingProps `json:"properties"`
	Timestamp  int64         `json:"timestamp"`
}

type TrackingProps struct {
	PhoneNumber      string          `json:"phone_number,omitempty"`
	UserID           int64           `json:"user_id,omitempty"`
	AccountID        string          `json:"account_id,omitempty"`
	ErrorCode        string          `json:"error_code,omitempty"`
	ErrorMessage     string          `json:"error_message,omitempty"`
	DeviceID         string          `json:"device_id,omitempty"`
	IPAddress        string          `json:"ip_address,omitempty"`
	CredentialType   string          `json:"credential_type,omitempty"`
	IsTrustedDevice  bool            `json:"is_trusted_device,omitempty"`
	VerificationType string          `json:"verification_type,omitempty"`
	FreezePeriodDays int             `json:"freeze_period_days,omitempty"`
	Custom           json.RawMessage `json:"custom,omitempty"`
}
