package screener

import (
	"context"
	"fmt"
	"strings"
	"time"

	"assiarius/internal/scraper"
)

// EnqueueScreenNews fetches the screener rows from screenURL (which should be a
// Finviz screener URL) and enqueues all quote-page news items within window.
//
// Items with empty links are skipped.
func EnqueueScreenNews(ctx context.Context, screenURL string, window time.Duration, q *NewsQueue) error {
	_ = ctx
	if q == nil {
		return fmt.Errorf("queue is nil")
	}
	if window <= 0 {
		return fmt.Errorf("window must be > 0")
	}

	// Ensure the URL includes a news time window filter.
	newsURL, err := BuildNewsScreenerURL(screenURL, window)
	if err != nil {
		return err
	}

	rows, err := FetchScreenRows(newsURL)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}

	now := time.Now()
	cutoff := now.Add(-window)

	for _, row := range rows {
		ticker := strings.TrimSpace(strings.ToUpper(row.Ticker))
		if ticker == "" {
			continue
		}

		items := scraper.FetchTickerNewsItem(ticker)
		if len(items) == 0 {
			continue
		}

		lastDate := time.Time{}
		for _, item := range items {
			item.Link = strings.TrimSpace(item.Link)
			if item.Link == "" {
				continue
			}

			ts, newLastDate, ok := parseFinvizNewsTimestamp(item.Time, lastDate, time.Local)
			lastDate = newLastDate
			if !ok {
				continue
			}

			// Items are returned newest->oldest; once we're older than cutoff, stop.
			if ts.Before(cutoff) {
				break
			}
			if ts.After(now.Add(2 * time.Minute)) {
				// Defensive: skip future timestamps.
				continue
			}

			q.Enqueue(NewsTask{Ticker: ticker, Item: item})
		}
	}

	return nil
}
