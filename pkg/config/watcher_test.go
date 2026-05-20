package config

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestNewWatcher(t *testing.T) {
	w := NewWatcher(nil, 30*time.Second)
	if w == nil {
		t.Fatal("expected non-nil watcher")
	}
	cfg := w.Get()
	if cfg != nil {
		t.Fatal("expected nil config before first load")
	}
}

func TestWatcherGetSet(t *testing.T) {
	w := NewWatcher(nil, 30*time.Second)
	initial := &Config{Values: map[string]string{"key": "val"}}
	w.store.Store(initial)
	got := w.Get()
	if got.Values["key"] != "val" {
		t.Fatalf("expected val, got %s", got.Values["key"])
	}
}

func TestWatcherPollAndPickup(t *testing.T) {
	mu := sync.Mutex{}
	items := map[string]string{
		"JWT_SECRET": "initial",
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Path[len("/internal/v1/config/items/"):]
		mu.Lock()
		val := items[code]
		mu.Unlock()
		resp := apiResponse{
			Code:    0,
			Message: "ok",
			Data: &configItem{
				Code:         code,
				CurrentValue: val,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	client := NewClient(ts.URL)
	w := NewWatcher(client, 100*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	updated := make(chan struct{}, 1)

	go w.Watch(ctx, func(c *Config) {
		if v, ok := c.Values["JWT_SECRET"]; ok && v == "updated" {
			select {
			case updated <- struct{}{}:
			default:
			}
		}
	})

	time.Sleep(250 * time.Millisecond)

	mu.Lock()
	items["JWT_SECRET"] = "updated"
	mu.Unlock()

	select {
	case <-updated:
	case <-ctx.Done():
		t.Fatal("timed out waiting for config update")
	}

	got := w.Get()
	if got.Values["JWT_SECRET"] != "updated" {
		t.Fatalf("expected updated, got %s", got.Values["JWT_SECRET"])
	}
}
