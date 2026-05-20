package model

import "time"

type ABTest struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Variants    []ABVariant `json:"variants"`
	Status      string      `json:"status"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at,omitempty"`
}

type ABVariant struct {
	Name   string  `json:"name"`
	Weight float64 `json:"weight"`
}

type ABAssignment struct {
	ExperimentID string `json:"experiment_id"`
	UserID       string `json:"user_id"`
	Variant      string `json:"variant"`
}

type ABEvent struct {
	ExperimentID string `json:"experiment_id"`
	UserID       string `json:"user_id"`
	Variant      string `json:"variant"`
	EventType    string `json:"event_type"`
	Timestamp    string `json:"timestamp"`
}

type ABVariantResult struct {
	Variant        string  `json:"variant"`
	Count          int     `json:"count"`
	Conversions    int     `json:"conversions"`
	ConversionRate float64 `json:"conversion_rate"`
	Confidence     float64 `json:"confidence"`
}

type ABTestResult struct {
	ExperimentID string            `json:"experiment_id"`
	Variants     []ABVariantResult `json:"variants"`
}
