package model

type SearchQuery struct {
	Keyword string `form:"q" json:"keyword"`
	Type    string `form:"type" json:"type,omitempty"`
	Page    int    `form:"page" json:"page,omitempty"`
	Size    int    `form:"size" json:"size,omitempty"`
}

type SearchResult struct {
	Type    string                 `json:"type"`
	ID      string                 `json:"id"`
	Title   string                 `json:"title"`
	Summary string                 `json:"summary,omitempty"`
	Meta    map[string]interface{} `json:"meta,omitempty"`
}

type SearchResponse struct {
	Query   string         `json:"query"`
	Results []SearchResult `json:"results"`
	Total   int            `json:"total"`
	Page    int            `json:"page"`
	Size    int            `json:"size"`
}

type QuickAction struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Icon        string `json:"icon,omitempty"`
	deeplink    string `json:"deeplink,omitempty"`
	Description string `json:"description,omitempty"`
}

type QuickActionsResponse struct {
	Tier    int           `json:"tier"`
	Actions []QuickAction `json:"actions"`
}
