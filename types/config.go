package types

import "strings"

type Config struct {
	DiscordToken   string
	DiscordAppID   string
	N8NAPIKey      string
	N8NBaseURL     string
	N8NWebhookPath string
	Prod           bool
}

func (c Config) N8NWebhookURL() string {
	basePath := "/webhook-test/"
	if c.Prod {
		basePath = "/webhook/"
	}

	baseURL := strings.TrimRight(strings.TrimSpace(c.N8NBaseURL), "/")
	webhookPath := strings.TrimLeft(strings.TrimSpace(c.N8NWebhookPath), "/")

	return baseURL + basePath + webhookPath
}
