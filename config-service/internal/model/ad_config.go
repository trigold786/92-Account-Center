package model

type AdConfig struct {
	Placement           string   `json:"placement"`
	PrimaryProvider     string   `json:"primary_provider"`
	BackupProvider      string   `json:"backup_provider"`
	ShowAds             bool     `json:"show_ads"`
	VideoMaxDurationSec int      `json:"video_max_duration_sec"`
	FrequencyPerHour    int      `json:"frequency_per_hour"`
	EnabledLevels       []string `json:"enabled_levels"`
}
