package health

type HealthResponse struct {
	Status   string                     `json:"status"`
	Checks   map[string]ComponentHealth `json:"checks,omitempty"`
}

func BuildResponse(checks map[string]ComponentHealth) HealthResponse {
	return BuildResponseConditional(checks, true)
}

func BuildResponseConditional(checks map[string]ComponentHealth, showDetails bool) HealthResponse {
	overall := "ok"
	for _, ch := range checks {
		if ch.Status == StatusDown {
			overall = "down"
			break
		}
		if ch.Status == StatusDegraded {
			overall = "degraded"
		}
	}
	resp := HealthResponse{Status: overall}
	if showDetails {
		resp.Checks = checks
	}
	return resp
}
