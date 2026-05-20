package service

import (
	"context"
	"testing"

	"github.com/trigold786/92-Account-Center/notification-service/internal/model"
)

func TestCreateTemplate(t *testing.T) {
	svc := NewTemplateService(nil)

	tpl, err := svc.CreateTemplate(context.Background(), &model.Template{
		Channel: "email",
		Name:    "welcome",
		Subject: "Welcome",
		Body:    "Hello {{name}}",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tpl.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
}

func TestListTemplates(t *testing.T) {
	svc := NewTemplateService(nil)

	svc.CreateTemplate(context.Background(), &model.Template{Channel: "email", Name: "t1"})
	svc.CreateTemplate(context.Background(), &model.Template{Channel: "sms", Name: "t2"})
	svc.CreateTemplate(context.Background(), &model.Template{Channel: "email", Name: "t3"})

	all, err := svc.ListTemplates(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("expected 3, got %d", len(all))
	}

	emailOnly, err := svc.ListTemplates(context.Background(), "email")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(emailOnly) != 2 {
		t.Fatalf("expected 2 email templates, got %d", len(emailOnly))
	}
}

func TestGetTemplate(t *testing.T) {
	svc := NewTemplateService(nil)

	tpl, _ := svc.CreateTemplate(context.Background(), &model.Template{Channel: "email", Name: "test"})

	got, err := svc.GetTemplate(context.Background(), tpl.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "test" {
		t.Fatalf("expected test, got %s", got.Name)
	}
}

func TestDeleteTemplate(t *testing.T) {
	svc := NewTemplateService(nil)

	tpl, _ := svc.CreateTemplate(context.Background(), &model.Template{Channel: "email", Name: "to-delete"})

	err := svc.DeleteTemplate(context.Background(), tpl.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	all, _ := svc.ListTemplates(context.Background(), "")
	if len(all) != 0 {
		t.Fatalf("expected 0 after delete, got %d", len(all))
	}
}
