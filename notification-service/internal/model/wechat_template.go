package model

type WeChatTemplate struct {
	TemplateType string `json:"template_type"`
	TemplateID   string `json:"template_id"`
	Title        string `json:"title"`
	Keywords     string `json:"keywords"`
}
