package screener

import (
	"context"
	"fmt"

	"assiarius/internal/scraper"
)

// EnqueueScreenTickers fetches a Finviz screener and enqueues tickers for LLM processing.
//
// It tries to pick up a volume column from the screener results if present; otherwise
// volumeText will be empty and the consumer can fetch it later.
func EnqueueScreenTickers(ctx context.Context, screen string, q *LLMQueue) error {
	_ = ctx
	if q == nil {
		return fmt.Errorf("queue is nil")
	}

	rows, err := FetchScreenRows(screen)
	if err != nil {
		return err
	}
	for _, row := range rows {
		volumeText := ""
		if stats, err := scraper.FetchTickerStatistics(row.Ticker); err == nil {
			if v, ok := stats.Text("Volume"); ok {
				volumeText = v
			}
		}
		q.Enqueue(row.Ticker, TickerSignals{
			VolumeText:  volumeText,
			VolumeValue: parseVolumeNumber(volumeText),
		})
	}

	return nil
}
