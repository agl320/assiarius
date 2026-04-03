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
	queue := screener.NewLLMQueue(llmClient, screener.GetQueueMinInterval())
	go queue.Start(ctx)

	// Enqueue once immediately so the first run doesn't wait a full interval.
	if err := screener.EnqueueScreenTickers(ctx, screen, queue); err != nil {
		return err
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	fmt.Println("Poll started...")
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			err := screener.EnqueueScreenTickers(ctx, screen, queue)
			if err != nil {
				return err
			}
		}
	}
}

