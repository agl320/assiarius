package screener

import (
	"fmt"

	"github.com/d3an/finviz/screener"
)

type ScreenRow struct {
	Ticker string
}

// FetchScreenRows fetches Finviz screener results and extracts a list of tickers.
func FetchScreenRows(screen string) ([]ScreenRow, error) {
	client := screener.New(nil)
	df, err := client.GetScreenerResults(screen)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch screener %q: %w", screen, err)
	}
	if df == nil || df.Nrow() == 0 {
		return nil, nil
	}

	records := df.Records()
	if len(records) == 0 {
		return nil, nil
	}

	const tickerIndex = 1

	rows := make([]ScreenRow, 0, len(records))
	for i, record := range records {
		if i == 0 {
			continue // header
		}
		if tickerIndex >= len(record) {
			continue
		}

		ticker := cleanTicker(record[tickerIndex])
		if ticker == "" {
			continue
		}

		rows = append(rows, ScreenRow{Ticker: ticker})
	}

	return rows, nil
}
