package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

type PersonalExport struct {
	UserID    int64                  `json:"user_id"`
	Profile   map[string]interface{} `json:"profile"`
	CreatedAt time.Time              `json:"created_at"`
	Encrypted bool                   `json:"encrypted"`
}

type ExportService struct {
	userRepo  interface{}
	cryptoKey string
}

func NewExportService(userRepo interface{}, cryptoKey string) *ExportService {
	return &ExportService{userRepo: userRepo, cryptoKey: cryptoKey}
}

func (s *ExportService) ExportPersonalData(ctx context.Context, userID int64) (*PersonalExport, error) {
	export := &PersonalExport{
		UserID:    userID,
		Profile:   map[string]interface{}{"id": userID, "exported_at": time.Now().Format(time.RFC3339)},
		CreatedAt: time.Now(),
		Encrypted: s.cryptoKey != "",
	}
	return export, nil
}

func (s *ExportService) RequestExport(ctx context.Context, userID int64) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate export request id: %w", err)
	}
	reqID := hex.EncodeToString(b)
	return reqID, nil
}

func (s *ExportService) ExportAdminReport(ctx context.Context, reportType string) ([]byte, error) {
	data := map[string]interface{}{
		"report_type":  reportType,
		"generated_at": time.Now().Format(time.RFC3339),
		"data":         []interface{}{},
	}
	return json.Marshal(data)
}
