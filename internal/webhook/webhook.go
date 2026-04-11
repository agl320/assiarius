package webhook

import (
	"context"
	"os"
)

func NotifyDiscord(ctx context.Context) error {
	url := os.Getenv("WEBHOOK_DISCORD_URL")

	if url == "" {
		return nil
	}

	
}