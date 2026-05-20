package health

import (
	"encoding/json"
	"testing"
)

func TestHealthResponseJSON(t *testing.T) {
	checks := map[string]ComponentHealth{
		"postgres": {Name: "postgres", Status: StatusUp, LatencyMs: 2},
	}
	resp := BuildResponse(checks)
	if resp.Status != "ok" {
		t.Fatalf("expected ok, got %s", resp.Status)
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	var parsed map[string]interface{}
	json.Unmarshal(data, &parsed)
	if parsed["checks"] == nil {
		t.Fatal("expected checks field in JSON")
	}
}
