package service

import (
	"context"
	"testing"
	"time"

	"github.com/trigold786/92-Account-Center/device-fingerprint-service/internal/model"
)

type mockDeviceRepo struct {
	devices map[string]*model.DeviceFingerprint
	nextID  uint64
}

func newMockDeviceRepo() *mockDeviceRepo {
	return &mockDeviceRepo{
		devices: make(map[string]*model.DeviceFingerprint),
		nextID:  1,
	}
}

func deviceKey(userID uint64, fingerprintID string) string {
	return string(rune(userID)) + ":" + fingerprintID
}

func (m *mockDeviceRepo) Save(ctx context.Context, fp *model.DeviceFingerprint) error {
	key := deviceKey(fp.UserID, fp.FingerprintID)
	if _, exists := m.devices[key]; !exists {
		fp.ID = m.nextID
		m.nextID++
	}
	m.devices[key] = fp
	return nil
}

func (m *mockDeviceRepo) GetByFingerprintID(ctx context.Context, userID uint64, fingerprintID string) (*model.DeviceFingerprint, error) {
	key := deviceKey(userID, fingerprintID)
	fp, ok := m.devices[key]
	if !ok {
		return nil, nil
	}
	return fp, nil
}

func (m *mockDeviceRepo) GetByUserID(ctx context.Context, userID uint64) ([]*model.DeviceFingerprint, error) {
	var result []*model.DeviceFingerprint
	for _, fp := range m.devices {
		if fp.UserID == userID {
			result = append(result, fp)
		}
	}
	return result, nil
}

func (m *mockDeviceRepo) Update(ctx context.Context, fp *model.DeviceFingerprint) error {
	key := deviceKey(fp.UserID, fp.FingerprintID)
	m.devices[key] = fp
	return nil
}

func (m *mockDeviceRepo) Delete(ctx context.Context, id uint64) error {
	for key, fp := range m.devices {
		if fp.ID == id {
			delete(m.devices, key)
			return nil
		}
	}
	return nil
}

func TestRegisterDevice_New(t *testing.T) {
	repo := newMockDeviceRepo()
	svc := NewDeviceFingerprintService(repo, 3, 0.3)

	req := &model.DeviceFingerprintRequest{
		FingerprintID: "fp-001",
		UserAgent:     "Mozilla/5.0",
		IPAddress:     "192.168.1.1",
		Country:       "CN",
		City:          "Beijing",
		Latitude:      39.9,
		Longitude:     116.4,
		Features:      []byte{1, 2, 3, 4},
	}

	resp, err := svc.RegisterDevice(context.Background(), 1, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.FingerprintID != "fp-001" {
		t.Errorf("expected fingerprint ID fp-001, got %s", resp.FingerprintID)
	}
	if resp.IsTrusted {
		t.Error("expected new device to not be trusted")
	}
}

func TestRegisterDevice_Existing(t *testing.T) {
	repo := newMockDeviceRepo()
	svc := NewDeviceFingerprintService(repo, 3, 0.3)

	req := &model.DeviceFingerprintRequest{
		FingerprintID: "fp-001",
		UserAgent:     "Mozilla/5.0",
		IPAddress:     "192.168.1.1",
		Country:       "CN",
		City:          "Beijing",
		Features:      []byte{1, 2, 3},
	}
	svc.RegisterDevice(context.Background(), 1, req)

	req2 := &model.DeviceFingerprintRequest{
		FingerprintID: "fp-001",
		UserAgent:     "Chrome/100",
		IPAddress:     "10.0.0.1",
		Country:       "US",
		City:          "NY",
		Features:      []byte{5, 6},
	}
	resp, err := svc.RegisterDevice(context.Background(), 1, req2)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.FingerprintID != "fp-001" {
		t.Errorf("expected fingerprint ID fp-001, got %s", resp.FingerprintID)
	}

	devices, _ := svc.GetUserDevices(context.Background(), 1)
	if len(devices) != 1 {
		t.Errorf("expected 1 device, got %d", len(devices))
	}
}

func TestTrustDevice(t *testing.T) {
	repo := newMockDeviceRepo()
	svc := NewDeviceFingerprintService(repo, 3, 0.3)

	req := &model.DeviceFingerprintRequest{
		FingerprintID: "fp-001",
		UserAgent:     "Mozilla/5.0",
		IPAddress:     "192.168.1.1",
		Country:       "CN",
		City:          "Beijing",
		Features:      []byte{1, 2, 3},
	}
	svc.RegisterDevice(context.Background(), 1, req)

	err := svc.TrustDevice(context.Background(), 1, "fp-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	trusted, err := svc.IsTrusted(context.Background(), 1, "fp-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !trusted {
		t.Error("expected device to be trusted")
	}
}

func TestTrustDevice_NotFound(t *testing.T) {
	repo := newMockDeviceRepo()
	svc := NewDeviceFingerprintService(repo, 3, 0.3)

	err := svc.TrustDevice(context.Background(), 1, "nonexistent")
	if err != ErrDeviceNotFound {
		t.Errorf("expected ErrDeviceNotFound, got %v", err)
	}
}

func TestIsTrusted_Trusted(t *testing.T) {
	repo := newMockDeviceRepo()
	svc := NewDeviceFingerprintService(repo, 3, 0.3)

	req := &model.DeviceFingerprintRequest{
		FingerprintID: "fp-001",
		UserAgent:     "Mozilla/5.0",
		IPAddress:     "192.168.1.1",
		Country:       "CN",
		City:          "Beijing",
		Features:      []byte{1, 2, 3},
	}
	svc.RegisterDevice(context.Background(), 1, req)
	svc.TrustDevice(context.Background(), 1, "fp-001")

	trusted, err := svc.IsTrusted(context.Background(), 1, "fp-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !trusted {
		t.Error("expected device to be trusted")
	}
}

func TestIsTrusted_Expired(t *testing.T) {
	repo := newMockDeviceRepo()
	svc := NewDeviceFingerprintService(repo, 3, 0.3)

	req := &model.DeviceFingerprintRequest{
		FingerprintID: "fp-001",
		UserAgent:     "Mozilla/5.0",
		IPAddress:     "192.168.1.1",
		Country:       "CN",
		City:          "Beijing",
		Features:      []byte{1, 2, 3},
	}
	svc.RegisterDevice(context.Background(), 1, req)
	svc.TrustDevice(context.Background(), 1, "fp-001")

	key := deviceKey(1, "fp-001")
	device := repo.devices[key]
	device.LastUsedAt = time.Now().Add(-4 * 24 * time.Hour).Unix()

	trusted, err := svc.IsTrusted(context.Background(), 1, "fp-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if trusted {
		t.Error("expected trust to be expired")
	}
}

func TestIsTrusted_NotTrusted(t *testing.T) {
	repo := newMockDeviceRepo()
	svc := NewDeviceFingerprintService(repo, 3, 0.3)

	req := &model.DeviceFingerprintRequest{
		FingerprintID: "fp-001",
		UserAgent:     "Mozilla/5.0",
		IPAddress:     "192.168.1.1",
		Country:       "CN",
		City:          "Beijing",
		Features:      []byte{1, 2, 3},
	}
	svc.RegisterDevice(context.Background(), 1, req)

	trusted, err := svc.IsTrusted(context.Background(), 1, "fp-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if trusted {
		t.Error("expected device to not be trusted")
	}
}

func TestIsTrusted_NotFound(t *testing.T) {
	repo := newMockDeviceRepo()
	svc := NewDeviceFingerprintService(repo, 3, 0.3)

	trusted, err := svc.IsTrusted(context.Background(), 1, "nonexistent")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if trusted {
		t.Error("expected not trusted for nonexistent device")
	}
}

func TestAssessRisk_GeoRisky(t *testing.T) {
	repo := newMockDeviceRepo()
	svc := NewDeviceFingerprintService(repo, 3, 0.3)

	req := &model.DeviceFingerprintRequest{
		FingerprintID: "fp-001",
		UserAgent:     "Mozilla/5.0",
		IPAddress:     "192.168.1.1",
		Country:       "CN",
		City:          "Beijing",
		Features:      []byte{1, 2, 3, 4},
	}
	svc.RegisterDevice(context.Background(), 1, req)

	req2 := &model.DeviceFingerprintRequest{
		FingerprintID: "fp-001",
		UserAgent:     "Mozilla/5.0",
		IPAddress:     "10.0.0.1",
		Country:       "US",
		City:          "NY",
		Features:      []byte{1, 2, 3, 4},
	}

	risky, err := svc.AssessRisk(context.Background(), 1, req2)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !risky {
		t.Error("expected geo location to be risky (different country)")
	}
}

func TestAssessRisk_FeatureRisky(t *testing.T) {
	repo := newMockDeviceRepo()
	svc := NewDeviceFingerprintService(repo, 3, 0.3)

	req := &model.DeviceFingerprintRequest{
		FingerprintID: "fp-001",
		UserAgent:     "Mozilla/5.0",
		IPAddress:     "192.168.1.1",
		Country:       "CN",
		City:          "Beijing",
		Features:      []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	}
	svc.RegisterDevice(context.Background(), 1, req)

	req2 := &model.DeviceFingerprintRequest{
		FingerprintID: "fp-001",
		UserAgent:     "Mozilla/5.0",
		IPAddress:     "192.168.1.1",
		Country:       "CN",
		City:          "Beijing",
		Features:      []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
	}

	risky, err := svc.AssessRisk(context.Background(), 1, req2)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !risky {
		t.Error("expected features to be risky (all bytes changed)")
	}
}

func TestAssessRisk_NotRisky(t *testing.T) {
	repo := newMockDeviceRepo()
	svc := NewDeviceFingerprintService(repo, 3, 0.3)

	req := &model.DeviceFingerprintRequest{
		FingerprintID: "fp-001",
		UserAgent:     "Mozilla/5.0",
		IPAddress:     "192.168.1.1",
		Country:       "CN",
		City:          "Beijing",
		Features:      []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	}
	svc.RegisterDevice(context.Background(), 1, req)

	req2 := &model.DeviceFingerprintRequest{
		FingerprintID: "fp-001",
		UserAgent:     "Mozilla/5.0",
		IPAddress:     "192.168.1.1",
		Country:       "CN",
		City:          "Beijing",
		Features:      []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	}

	risky, err := svc.AssessRisk(context.Background(), 1, req2)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if risky {
		t.Error("expected no risk for same device and features")
	}
}

func TestAssessRisk_NewDevice(t *testing.T) {
	repo := newMockDeviceRepo()
	svc := NewDeviceFingerprintService(repo, 3, 0.3)

	req := &model.DeviceFingerprintRequest{
		FingerprintID: "fp-new",
		UserAgent:     "Mozilla/5.0",
		IPAddress:     "192.168.1.1",
		Country:       "CN",
		City:          "Beijing",
		Features:      []byte{1, 2, 3},
	}

	risky, err := svc.AssessRisk(context.Background(), 1, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !risky {
		t.Error("expected new device to be risky")
	}
}
