package screener

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"assiarius/internal/llm"
	"assiarius/internal/scraper"
)

type NewsItem = scraper.NewsItem

const (
	VerdictVeryPositive = "VERY POSITIVE"
	VerdictPositive     = "POSITIVE"
	VerdictNeutral      = "NEUTRAL"
	VerdictNegative     = "NEGATIVE"
	VerdictVeryNegative = "VERY NEGATIVE"
	VerdictUndetermined = "UNDETERMINED"
)

func RunScreen(ctx context.Context, screen string, includeNews bool, llmClient llm.Client) error {
	rows, err := FetchScreenRows(screen)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		fmt.Printf("No results found for screener %q\n", screen)
		return nil
	}

	if includeNews {
		extractNewsSlice(ctx, rows, llmClient)
		return nil
	}

	printTickers(rows)
	return nil
}

func printTickers(rows []ScreenRow) {
	for i, row := range rows {
		volume, _ := scraper.GetTickerValueAny(row.Ticker, "Volume", "Avg Volume")
		if volume != "" {
			fmt.Printf("%d %s Volume: %s\n", i+1, row.Ticker, volume)
		} else {
			fmt.Printf("%d %s\n", i+1, row.Ticker)
		}
	}
}

func extractNewsSlice(ctx context.Context, rows []ScreenRow, llmClient llm.Client) {
	for _, row := range rows {
		GetNewsForTicker(ctx, row.Ticker, llmClient)
	}
}

func cleanTicker(s string) string {
	s = strings.ToUpper(s)

	re := regexp.MustCompile(`[^A-Z0-9]`)
	s = re.ReplaceAllString(s, "")

	return s
}

func GetNewsForTicker(ctx context.Context, ticker string, llmClient llm.Client) []NewsItem {
	newsItems, _ := GetNewsForTickerWithVolume(ctx, ticker, "", llmClient)
	return newsItems
}

func GetNewsForTickerWithVolume(
	ctx context.Context,
	ticker string,
	volumeText string,
	llmClient llm.Client,
) ([]NewsItem, string) {
	volume := strings.TrimSpace(volumeText)

	// If we need both news and volume, fetch the quote page once.
	stats, newsItems, err := scraper.FetchTickerQuote(ticker)
	if err != nil {
		newsItems = scraper.FetchTickerNewsItem(ticker)
		if volume == "" {
			volume, _ = scraper.GetTickerValueAny(ticker, "Volume", "Avg Volume")
		}
	} else if volume == "" {
		if v, ok := stats.Get("Volume"); ok && !v.Missing() {
			volume = strings.TrimSpace(v.Raw)
		}
		if volume == "" {
			if v, ok := stats.Get("Avg Volume"); ok && !v.Missing() {
				volume = strings.TrimSpace(v.Raw)
			}
		}
	}

	if len(newsItems) == 0 {
		return newsItems, volume
	}

	item := newsItems[0]

	timeStr := strings.TrimSpace(strings.Join(strings.Fields(item.Time), " "))
	text, _ := scraper.ExtractNewsFromLink(item.Link)

	verdict, err := getVerdictFromGemini(ctx, text, llmClient)
	if err != nil {
		verdict = VerdictUndetermined
	}

	// Build output dynamically
	parts := []string{ticker}

	if timeStr != "" {
		parts = append(parts, "Time:", timeStr)
	}
	if volume != "" {
		parts = append(parts, "Volume:", volume)
	}

	parts = append(parts, "Verdict:", verdict)

	fmt.Println(strings.Join(parts, " "))

	return newsItems, volume
}

func normalizeVerdict(raw string) string {
	if raw == "" {
		return VerdictUndetermined
	}

	// Normalize to make substring checks reliable.
	normalized := strings.ToUpper(strings.Join(strings.Fields(raw), " "))

	// Check more-specific phrases before less-specific ones.
	switch {
	case strings.Contains(normalized, VerdictVeryPositive):
		return VerdictVeryPositive
	case strings.Contains(normalized, VerdictVeryNegative):
		return VerdictVeryNegative
	case strings.Contains(normalized, VerdictNeutral):
		return VerdictNeutral
	case strings.Contains(normalized, VerdictPositive):
		return VerdictPositive
	case strings.Contains(normalized, VerdictNegative):
		return VerdictNegative
	case strings.Contains(normalized, VerdictUndetermined):
		return VerdictUndetermined
	default:
		return VerdictUndetermined
	}
}

// getVerdictFromGemini sends the article text to Gemini and retrieves a sentiment verdict.
func getVerdictFromGemini(ctx context.Context, text string, llmClient llm.Client) (string, error) {
	if llmClient == nil {
		return "", fmt.Errorf("LLM client is not configured")
	}

	result, err := llmClient.Process(ctx, llm.Prompt{
		Prompt:  "Determine if the news article is VERY POSITIVE, POSITIVE, NEUTRAL, NEGATIVE, VERY NEGATIVE or UNDETERMINED for the stock mentioned. Only respond with a single verdict.",
		Message: text,
	})
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return VerdictUndetermined, nil
		}
		return "", err
	}
	return normalizeVerdict(result), nil
}

