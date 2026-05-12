package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/trigold786/92-Account-Center/kyb-service/internal/model"
	"github.com/trigold786/92-Account-Center/kyb-service/pkg/crypto"
)

type mockEnterpriseRepo struct {
	enterprises map[uuid.UUID]*model.Enterprise
	createErr   error
}

func newMockRepo() *mockEnterpriseRepo {
	return &mockEnterpriseRepo{enterprises: make(map[uuid.UUID]*model.Enterprise)}
}

func (m *mockEnterpriseRepo) Create(_ context.Context, e *model.Enterprise) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.enterprises[e.EnterpriseID] = e
	return nil
}

func (m *mockEnterpriseRepo) GetByID(_ context.Context, id uuid.UUID) (*model.Enterprise, error) {
	return m.enterprises[id], nil
}

func (m *mockEnterpriseRepo) GetByUserID(_ context.Context, _ uuid.UUID) ([]*model.Enterprise, error) {
	return nil, nil
}

func (m *mockEnterpriseRepo) Update(_ context.Context, e *model.Enterprise) error {
	m.enterprises[e.EnterpriseID] = e
	return nil
}

func validUUID() string {
	return uuid.New().String()
}

func TestSubmitEnterpriseInfo_Success(t *testing.T) {
	repo := newMockRepo()
	key := make([]byte, 16)
	rand.Read(key)
	svc := NewKYBService(repo, key)

	userID := validUUID()
	req := &model.EnterpriseInfoRequest{
		UserID:                  userID,
		CompanyName:             "Test Corp",
		UnifiedSocialCreditCode: "91110000MA000000XX",
		LegalPersonName:         "Zhang San",
		LegalPersonIDNumber:     "110101199001011234",
		BankName:                "ICBC",
		BankAccountNumber:       "6222021234567890123",
	}

	resp, err := svc.SubmitEnterpriseInfo(context.Background(), userID, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.EnterpriseID == "" {
		t.Error("expected enterprise ID")
	}
	if resp.VerificationStatus != "pending" {
		t.Errorf("expected pending status, got %s", resp.VerificationStatus)
	}

	if len(repo.enterprises) != 1 {
		t.Fatalf("expected 1 enterprise, got %d", len(repo.enterprises))
	}
	for _, e := range repo.enterprises {
		if _, err := hex.DecodeString(e.LegalPersonIDNumber); err != nil {
			t.Errorf("legal person ID should be hex-encoded ciphertext: %v", err)
		}
		if _, err := hex.DecodeString(e.BankAccountNumber); err != nil {
			t.Errorf("bank account should be hex-encoded ciphertext: %v", err)
		}
	}
}

func TestSubmitEnterpriseInfo_InvalidUserID(t *testing.T) {
	repo := newMockRepo()
	key := make([]byte, 16)
	rand.Read(key)
	svc := NewKYBService(repo, key)

	req := &model.EnterpriseInfoRequest{
		UserID:                  "not-a-uuid",
		CompanyName:             "Test",
		UnifiedSocialCreditCode: "123",
		LegalPersonName:         "Test",
		LegalPersonIDNumber:     "123",
		BankName:                "Test",
		BankAccountNumber:       "123",
	}

	_, err := svc.SubmitEnterpriseInfo(context.Background(), "not-a-uuid", req)
	if err == nil {
		t.Fatal("expected error for invalid user ID")
	}
}

func TestInitiateMicroPayment_Success(t *testing.T) {
	repo := newMockRepo()
	key := make([]byte, 16)
	rand.Read(key)
	svc := NewKYBService(repo, key)

	eid := uuid.New()
	enterprise := &model.Enterprise{
		EnterpriseID:       eid,
		UserID:             uuid.New(),
		MicroPaymentStatus: model.MicroPaymentStatusPending,
		BankName:           "ICBC",
		BankAccountNumber:  "6222021234567890123",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
	repo.enterprises[eid] = enterprise

	resp, err := svc.InitiateMicroPayment(context.Background(), eid.String())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.EnterpriseID != eid.String() {
		t.Errorf("expected enterprise ID %s, got %s", eid.String(), resp.EnterpriseID)
	}
	if resp.Amount < 0.01 || resp.Amount > 0.1 {
		t.Errorf("expected amount between 0.01 and 0.1, got %f", resp.Amount)
	}
	if resp.BankName != "ICBC" {
		t.Errorf("expected bank name ICBC, got %s", resp.BankName)
	}
}

func TestInitiateMicroPayment_InvalidEnterpriseID(t *testing.T) {
	repo := newMockRepo()
	key := make([]byte, 16)
	rand.Read(key)
	svc := NewKYBService(repo, key)

	_, err := svc.InitiateMicroPayment(context.Background(), "invalid-uuid")
	if err == nil {
		t.Fatal("expected error for invalid enterprise ID")
	}
}

func TestInitiateMicroPayment_NotFound(t *testing.T) {
	repo := newMockRepo()
	key := make([]byte, 16)
	rand.Read(key)
	svc := NewKYBService(repo, key)

	_, err := svc.InitiateMicroPayment(context.Background(), uuid.New().String())
	if err != ErrEnterpriseNotFound {
		t.Fatalf("expected ErrEnterpriseNotFound, got %v", err)
	}
}

func TestVerifyMicroPayment_CorrectAmount(t *testing.T) {
	repo := newMockRepo()
	key := make([]byte, 16)
	rand.Read(key)
	svc := NewKYBService(repo, key)

	eid := uuid.New()
	enterprise := &model.Enterprise{
		EnterpriseID:         eid,
		UserID:               uuid.New(),
		MicroPaymentStatus:   model.MicroPaymentStatusPending,
		MicroPaymentAmount:   0.05,
		FaceVerificationStatus: model.FaceVerificationStatusPending,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}
	repo.enterprises[eid] = enterprise

	err := svc.VerifyMicroPayment(context.Background(), eid.String(), 0.05)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.enterprises[eid].MicroPaymentStatus != model.MicroPaymentStatusVerified {
		t.Error("expected micro payment status to be verified")
	}
}

func TestVerifyMicroPayment_IncorrectAmount(t *testing.T) {
	repo := newMockRepo()
	key := make([]byte, 16)
	rand.Read(key)
	svc := NewKYBService(repo, key)

	eid := uuid.New()
	enterprise := &model.Enterprise{
		EnterpriseID:       eid,
		UserID:             uuid.New(),
		MicroPaymentStatus: model.MicroPaymentStatusPending,
		MicroPaymentAmount: 0.05,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
	repo.enterprises[eid] = enterprise

	err := svc.VerifyMicroPayment(context.Background(), eid.String(), 0.09)
	if err != nil {
		t.Fatalf("expected no error from VerifyMicroPayment, got %v", err)
	}
	if repo.enterprises[eid].MicroPaymentStatus != model.MicroPaymentStatusFailed {
		t.Error("expected micro payment status to be failed")
	}
}

func TestEncryptDecrypt_Roundtrip(t *testing.T) {
	key := make([]byte, 16)
	rand.Read(key)

	plaintext := "sensitive-data-12345"

	ciphertext, err := crypto.Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	matched, _ := regexp.MatchString(`^[0-9a-f]+$`, ciphertext)
	if !matched {
		t.Errorf("expected hex output, got %s", ciphertext)
	}

	decrypted, err := crypto.Decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}
	if decrypted != plaintext {
		t.Errorf("expected %s, got %s", plaintext, decrypted)
	}
}

func TestEncrypt_InvalidKeySize(t *testing.T) {
	_, err := crypto.Encrypt("test", []byte("short"))
	if err == nil {
		t.Error("expected error for invalid key size")
	}
}

func TestDecrypt_InvalidKeySize(t *testing.T) {
	_, err := crypto.Decrypt("abcdef", []byte("short"))
	if err == nil {
		t.Error("expected error for invalid key size")
	}
}

func TestMaskBankAccount(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1234567890", "****7890"},
		{"1234", "1234"},
		{"123", "123"},
		{"1234567890123456", "****3456"},
	}
	for _, tt := range tests {
		result := maskBankAccount(tt.input)
		if result != tt.expected {
			t.Errorf("maskBankAccount(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestGenerateMicroPaymentAmount(t *testing.T) {
	for i := 0; i < 100; i++ {
		amount, err := generateMicroPaymentAmount()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if amount < 0.01 || amount > 0.1 {
			t.Errorf("amount %f out of range [0.01, 0.1]", amount)
		}
		scaled := int(amount * 100)
		if float64(scaled)/100.0 != amount {
			t.Errorf("amount %f has more than 2 decimal places", amount)
		}
	}
}
