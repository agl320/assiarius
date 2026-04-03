package screener

import (
	"context"
	"fmt"
	"strings"

	"github.com/d3an/finviz/screener"
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

	client := screener.New(nil)
	df, err := client.GetScreenerResults(screen)
	if err != nil {
		return fmt.Errorf("failed to fetch screener %q: %w", screen, err)
	}
	if df == nil || df.Nrow() == 0 {
		return nil
	}

	records := df.Records()
	if len(records) == 0 {
		return nil
	}

	tickerIndex := guessColumnIndex(records, func(col string) bool {
		col = strings.TrimSpace(strings.ToUpper(col))
		return col == "TICKER" || col == "SYMBOL"
	}, 1)

	volumeIndex := guessColumnIndex(records, func(col string) bool {
		col = strings.TrimSpace(strings.ToUpper(col))
		return col == "VOLUME" || strings.Contains(col, "VOLUME")
	}, -1)

	for i, record := range records {
		if i == 0 {
			continue // header
		}
		if tickerIndex < 0 || tickerIndex >= len(record) {
			continue
		}

		ticker := cleanTicker(record[tickerIndex])
		if ticker == "" {
			continue
		}

		volumeText := ""
		if volumeIndex >= 0 && volumeIndex < len(record) {
			volumeText = strings.TrimSpace(record[volumeIndex])
		}

		q.Enqueue(ticker, volumeText)
	}

	return nil
}

func guessColumnIndex(records [][]string, match func(col string) bool, fallback int) int {
	if len(records) == 0 {
		return fallback
	}
	header := records[0]
	for i, name := range header {
		if match(name) {
			return i
		}
	}
	return fallback
}
