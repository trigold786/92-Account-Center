package service

import (
	"context"
	"testing"
)

func TestExportPersonalData(t *testing.T) {
	svc := NewExportService(nil, "")
	userID := int64(42)
	export, err := svc.ExportPersonalData(context.Background(), userID)
	if err != nil {
		t.Fatalf("ExportPersonalData failed: %v", err)
	}
	if export.UserID != userID {
		t.Fatalf("unexpected user ID: %d", export.UserID)
	}
	if export.Encrypted {
		t.Log("export is encrypted")
	}
}

func TestExportRequestFlow(t *testing.T) {
	svc := NewExportService(nil, "")
	reqID, err := svc.RequestExport(context.Background(), 42)
	if err != nil {
		t.Fatalf("RequestExport failed: %v", err)
	}
	if reqID == "" {
		t.Fatal("expected non-empty request ID")
	}
}
