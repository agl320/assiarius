package poll

import (
	"assiarius/internal/llm"
	"assiarius/internal/screener"
	"context"
	"fmt"
	"time"
)

type ScreenerResult struct {
	Ticker string
	Price  float64
}

func StartPoller(ctx context.Context, screen string, interval time.Duration, llmClient llm.Client) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	fmt.Println("Poll started...")
	for {
		// wait for channel
		<-ticker.C
		err := screener.RunScreen(ctx, screen, true, llmClient)
		if err != nil {
			return err
		}
	}
}

