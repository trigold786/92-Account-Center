package health

type HealthResponse struct {
	Status   string                       `json:"status"`
	Checks   map[string]ComponentHealth   `json:"checks,omitempty"`
}

func BuildResponse(checks map[string]ComponentHealth) HealthResponse {
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
	return HealthResponse{
		Status: overall,
		Checks: checks,
	}
}
