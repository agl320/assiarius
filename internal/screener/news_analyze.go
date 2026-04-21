package screener

import (
	"context"
	"log"
	"strings"

	"assiarius/internal/llm"
	"assiarius/internal/scraper"
)

// AnalyzeNewsItem extracts text from a specific news item link and gets an LLM sentiment verdict.
//
// If extraction or LLM fails, Verdict will be UNDETERMINED.
func AnalyzeNewsItem(
	ctx context.Context,
	ticker string,
	item scraper.NewsItem,
	signals TickerSignals,
	llmClient llm.Client,
) (ScreenResult, error) {
	volume := strings.TrimSpace(signals.VolumeText)

	text, _ := scraper.ExtractNewsFromLink(item.Link)
	if strings.TrimSpace(text) == "" {
		log.Printf("llm: empty extracted text ticker=%s link=%s headline=%q", strings.TrimSpace(strings.ToUpper(ticker)), strings.TrimSpace(item.Link), strings.TrimSpace(item.Headline))
		return ScreenResult{Ticker: ticker, Volume: volume, Verdict: VerdictUndetermined, Latest: item}, nil
	}
	verdict, err := getVerdictFromGemini(ctx, text, llmClient)
	if err != nil {
		log.Printf("llm: error ticker=%s link=%s headline=%q err=%v", strings.TrimSpace(strings.ToUpper(ticker)), strings.TrimSpace(item.Link), strings.TrimSpace(item.Headline), err)
		verdict = VerdictUndetermined
		// Return nil error so the queue can continue without stopping; this matches
		// the "best effort" nature of polling.
		return ScreenResult{Ticker: ticker, Volume: volume, Verdict: verdict, Latest: item}, nil
	}

	return ScreenResult{Ticker: ticker, Volume: volume, Verdict: verdict, Latest: item}, nil
}
