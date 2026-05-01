package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"tradingcorpbot/types"
)

func LoadConfig() (types.Config, error) {
	if err := loadDotEnv(".env"); err != nil {
		return types.Config{}, err
	}

	discordToken, err := requiredValue("DISCORD_BOT_TOKEN")
	if err != nil {
		return types.Config{}, err
	}

	discordAppID, err := requiredValue("DISCORD_APP_ID")
	if err != nil {
		return types.Config{}, err
	}

	n8nAPIKey, err := requiredValue("N8N_API_KEY")
	if err != nil {
		return types.Config{}, err
	}

	n8nWebhookPath, err := requiredValue("N8N_WEBHOOK_PATH")
	if err != nil {
		return types.Config{}, err
	}

	baseURL := strings.TrimRight(getEnv("N8N_BASE_URL", "https://n8n.zastial.com"), "/")
	prod := parseBool(os.Getenv("N8N_PROD"))

	return types.Config{
		DiscordToken:   discordToken,
		DiscordAppID:   discordAppID,
		N8NAPIKey:      n8nAPIKey,
		N8NBaseURL:     baseURL,
		N8NWebhookPath: strings.TrimLeft(strings.TrimSpace(n8nWebhookPath), "/"),
		Prod:           prod,
	}, nil
}

func requiredValue(name string) (string, error) {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value, nil
	}

	secretValue, err := readSecret(name)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("missing required environment variable or secret: %s", name)
		}

		return "", err
	}

	secretValue = strings.TrimSpace(secretValue)
	if secretValue == "" {
		return "", fmt.Errorf("missing required environment variable or secret: %s", name)
	}

	return secretValue, nil
}

func readSecret(name string) (string, error) {
	paths := []string{
		"/run/secrets/" + name,
		"/run/secrets/" + strings.ToLower(name),
	}

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err == nil {
			return string(data), nil
		}

		if !isNotExistError(err) {
			return "", fmt.Errorf("read secret %s: %w", name, err)
		}
	}

	return "", os.ErrNotExist
}

func isNotExistError(err error) bool {
	return err != nil && (errors.Is(err, os.ErrNotExist) || os.IsNotExist(err))
}

func loadDotEnv(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = stripQuotes(strings.TrimSpace(value))
		if key == "" {
			continue
		}

		if _, exists := os.LookupEnv(key); !exists {
			if err := os.Setenv(key, value); err != nil {
				return fmt.Errorf("set env %s: %w", key, err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func stripQuotes(value string) string {
	if len(value) < 2 {
		return value
	}

	if (strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`)) || (strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
		return value[1 : len(value)-1]
	}

	return value
}

func requiredEnv(name string) (string, error) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return "", fmt.Errorf("missing required environment variable: %s", name)
	}

	return value, nil
}

func getEnv(name, fallback string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}

	return value
}

func parseBool(value string) bool {
	parsed, err := strconv.ParseBool(strings.TrimSpace(value))
	return err == nil && parsed
}
