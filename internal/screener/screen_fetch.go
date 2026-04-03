package screener

import (
	"fmt"
	"strings"

	"github.com/d3an/finviz/screener"
)

type ScreenRow struct {
	Ticker     string
	VolumeText string
}

// FetchScreenRows fetches Finviz screener results and extracts a list of tickers.
//
// It also attempts to extract a volume-like column if present in the screener output.
// If no volume column exists, VolumeText will be empty.
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

	tickerIndex := guessColumnIndex(records, func(col string) bool {
		col = strings.TrimSpace(strings.ToUpper(col))
		return col == "TICKER" || col == "SYMBOL"
	}, 1)

	volumeIndex := guessColumnIndex(records, func(col string) bool {
		col = strings.TrimSpace(strings.ToUpper(col))
		return col == "VOLUME" || strings.Contains(col, "VOLUME")
	}, -1)

	rows := make([]ScreenRow, 0, len(records))
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

		rows = append(rows, ScreenRow{Ticker: ticker, VolumeText: volumeText})
	}

	return rows, nil
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
