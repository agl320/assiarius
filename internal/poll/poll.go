package poll

import (
	"assiarius/internal/llm"
	"assiarius/internal/screener"
	"context"
	"time"
)

// StartPoller starts a screener poller and returns a channel of per-ticker results.
//
// The poller stops when ctx is cancelled.
func StartPoller(ctx context.Context, screen string, interval time.Duration, llmClient llm.Client) (<-chan screener.ScreenResult, error) {
	queue := screener.NewLLMQueue(llmClient, screener.GetQueueMinInterval())
	go queue.Start(ctx)

	// Enqueue once immediately so the first run doesn't wait a full interval.
	if err := screener.EnqueueScreenTickers(ctx, screen, queue); err != nil {
		return nil, err
	}

	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_ = screener.EnqueueScreenTickers(ctx, screen, queue)
			}
		}
	}()

	return queue.Results(), nil
}

