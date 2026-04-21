package screener

import (
	"context"
	"fmt"
	"log"
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
	log.Printf("poll: screener url=%s", newsURL)

	rows, err := FetchScreenRows(newsURL)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		log.Printf("poll: screener returned 0 rows")
		return nil
	}

	now := time.Now()
	cutoff := now.Add(-window)

	enqueued := 0
	skippedDup := 0
	skippedEmptyLink := 0
	scannedItems := 0

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
				skippedEmptyLink++
				continue
			}
			scannedItems++

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

			if q.Enqueue(NewsTask{Ticker: ticker, Item: item}) {
				enqueued++
			} else {
				skippedDup++
			}
		}
	}

	log.Printf("poll: enqueue done tickers=%d items_scanned=%d enqueued=%d skipped_dup=%d skipped_empty_link=%d", len(rows), scannedItems, enqueued, skippedDup, skippedEmptyLink)

	return nil
}
