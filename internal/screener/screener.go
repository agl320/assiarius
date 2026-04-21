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

// TickerSignals are optional hints that may already be known when a ticker is
// enqueued (e.g. from a screener row). 
type TickerSignals struct {
	VolumeText  string
	VolumeValue float64
}

type ScreenResult struct {
	Ticker  string
	Volume  string
	Verdict string
	Latest  NewsItem
}

type ScreenTicker struct {
	Ticker string
	Volume string
}

type ScreenRun struct {
	Screen  string
	Tickers []ScreenTicker
	Results []ScreenResult
}

const (
	VerdictVeryPositive = "VERY POSITIVE"
	VerdictPositive     = "POSITIVE"
	VerdictNeutral      = "NEUTRAL"
	VerdictNegative     = "NEGATIVE"
	VerdictVeryNegative = "VERY NEGATIVE"
	VerdictUndetermined = "UNDETERMINED"
)

func RunScreen(ctx context.Context, screen string, includeNews bool, llmClient llm.Client) (ScreenRun, error) {
	rows, err := FetchScreenRows(screen)
	if err != nil {
		return ScreenRun{}, err
	}

	run := ScreenRun{Screen: screen}
	if len(rows) == 0 {
		return run, nil
	}

	run.Tickers = make([]ScreenTicker, 0, len(rows))
	for _, row := range rows {
		run.Tickers = append(run.Tickers, ScreenTicker{Ticker: row.Ticker})
	}

	if !includeNews {
		populateTickerVolumes(run.Tickers)
		return run, nil
	}

	run.Results = make([]ScreenResult, 0, len(rows))
	for _, row := range rows {
		res, err := AnalyzeLatestNews(ctx, row.Ticker, TickerSignals{}, llmClient)
		if err != nil {
			continue
		}
		run.Results = append(run.Results, res)
	}

	return run, nil
}

func populateTickerVolumes(tickers []ScreenTicker) {
	for i := range tickers {
		volume := ""
		if stats, err := scraper.FetchTickerStatistics(tickers[i].Ticker); err == nil {
			if v, ok := stats.Text("Volume"); ok {
				volume = v
			}
		}
		tickers[i].Volume = volume
	}
}

func cleanTicker(s string) string {
	s = strings.ToUpper(s)

	re := regexp.MustCompile(`[^A-Z0-9]`)
	s = re.ReplaceAllString(s, "")

	return s
}

// GetNewsForTicker analyzes the latest news item for the ticker.
func GetNewsForTicker(ctx context.Context, ticker string, llmClient llm.Client) (ScreenResult, error) {
	return AnalyzeLatestNews(ctx, ticker, TickerSignals{}, llmClient)
}

// AnalyzeLatestNews fetches the latest news item for a ticker, extracts its text,
// and gets an LLM sentiment verdict.
func AnalyzeLatestNews(
	ctx context.Context,
	ticker string,
	signals TickerSignals,
	llmClient llm.Client,
) (ScreenResult, error) {
	volume := strings.TrimSpace(signals.VolumeText)

	// If we need both news and volume, fetch the quote page once.
	stats, newsItems, err := scraper.FetchTickerQuote(ticker)
	if err != nil {
		newsItems = scraper.FetchTickerNewsItem(ticker)
	} else if volume == "" {
		if v, ok := stats.Text("Volume"); ok {
			volume = v
		}
	}

	if len(newsItems) == 0 {
		return ScreenResult{Ticker: ticker, Volume: volume, Verdict: VerdictUndetermined}, nil
	}

	item := newsItems[0]

	timeStr := strings.TrimSpace(strings.Join(strings.Fields(item.Time), " "))
	text, _ := scraper.ExtractNewsFromLink(item.Link)

	verdict, err := getVerdictFromGemini(ctx, text, llmClient)
	if err != nil {
		verdict = VerdictUndetermined
	}

	_ = timeStr // preserved for formatting via NewsItem fields
	return ScreenResult{
		Ticker:  ticker,
		Volume:  volume,
		Verdict: verdict,
		Latest:  item,
	}, nil
}

func FormatScreenResult(res ScreenResult) string {
	parts := []string{strings.TrimSpace(strings.ToUpper(res.Ticker))}

	if t := strings.TrimSpace(strings.Join(strings.Fields(res.Latest.Time), " ")); t != "" {
		parts = append(parts, "Time:", t)
	}
	if v := strings.TrimSpace(res.Volume); v != "" {
		parts = append(parts, "Volume:", v)
	}
	if verdict := strings.TrimSpace(res.Verdict); verdict != "" {
		parts = append(parts, "Verdict:", verdict)
	}

	return strings.Join(parts, " ")
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

