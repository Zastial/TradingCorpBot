package n8n

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"

	bottypes "tradingcorpbot/types"
)

func SendRequest(ctx context.Context, cfg bottypes.Config, ticker string, user *discordgo.User) error {
	payload := map[string]string{
		"ticker": ticker,
		"user":   "",
		"userId": "",
	}

	if user != nil {
		payload["user"] = user.Username
		payload["userId"] = user.ID
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.N8NWebhookURL(), bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", cfg.N8NAPIKey)

	client := &http.Client{Timeout: 20 * time.Second}
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("n8n webhook returned %s", response.Status)
	}

	return nil
}
