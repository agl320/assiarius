package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type discordPayload struct {
	Content string `json:"content"`
}

func NotifyDiscord(ctx context.Context, content string) error {
	url := strings.TrimSpace(os.Getenv("WEBHOOK_DISCORD_URL"))
	if url == "" {
		return nil
	}

	body, err := json.Marshal(discordPayload{Content: content})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("discord webhook returned %s", resp.Status)
	}
	return nil
}