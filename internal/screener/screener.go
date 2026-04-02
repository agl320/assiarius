package screener

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/d3an/finviz/screener"
	"github.com/go-gota/gota/dataframe"

	"assiarius/internal/llm"
	"assiarius/internal/scraper"
)

type NewsItem struct {
	Headline string
	Link     string
	Time     string
}

const (
	VerdictVeryPositive = "VERY POSITIVE"
	VerdictPositive     = "POSITIVE"
	VerdictNeutral      = "NEUTRAL"
	VerdictNegative     = "NEGATIVE"
	VerdictVeryNegative = "VERY NEGATIVE"
	VerdictUndetermined = "UNDETERMINED"
)

func RunScreen(ctx context.Context, screen string, includeNews bool, llmClient llm.Client) error {
	client := screener.New(nil)
	df, err := client.GetScreenerResults(screen)
	if err != nil {
		return fmt.Errorf("failed to fetch screener %q: %w", screen, err)
	}

	if df == nil || df.Nrow() == 0 {
		fmt.Printf("No results found for screener %q\n", screen)
		return nil
	}

	if includeNews {
		extractNewsSlice(ctx, df, llmClient)
		return nil
	}

	printTickers(df)
	return nil
}

func printTickers(df *dataframe.DataFrame) {
	records := df.Records()
	for index, record := range records {
		if index == 0 {
			// Header row from gota dataframe.
			continue
		}
		if len(record) > 1 {
			ticker := cleanTicker(record[1])
			if ticker == "" {
				continue
			}
			volume, err := scraper.GetTickerVolume(ticker)
			if err == nil && volume != "" {
				fmt.Printf("%d %s Volume: %s\n", index, ticker, volume)
			} else {
				fmt.Printf("%d %s\n", index, ticker)
			}
		}
	}
}

func extractNewsSlice(ctx context.Context, df *dataframe.DataFrame, llmClient llm.Client) {
	records := df.Records()

	for index, record := range records {
		if index == 0 {
			// Header row from gota dataframe.
			continue
		}
		if len(record) > 1 {
			ticker := cleanTicker(record[1])
			if ticker == "" {
				continue
			}
			GetNewsForTicker(ctx, ticker, llmClient)
		}
	}
	
}

func cleanTicker(s string) string {
	s = strings.ToUpper(s)

	re := regexp.MustCompile(`[^A-Z0-9]`)
	s = re.ReplaceAllString(s, "")

	return s
}

func GetNewsForTicker(ctx context.Context, ticker string, llmClient llm.Client) []NewsItem {
	newsItems := FetchTickerNewsItem(ticker)
	volume, _ := scraper.GetTickerVolume(ticker)

	for i := 0; i < len(newsItems) && i < 1; i++ {
		item := newsItems[i]
		timeStr := strings.TrimSpace(strings.Join(strings.Fields(item.Time), " "))
		text, _ := scraper.ExtractNewsFromLink(item.Link)

		verdict, err := getVerdictFromGemini(ctx, text, llmClient)
		if err != nil {
			if volume != "" {
				if timeStr != "" {
					fmt.Printf("%s Time: %s Volume: %s Verdict: %s\n", ticker, timeStr, volume, VerdictUndetermined)
				} else {
					fmt.Printf("%s Volume: %s Verdict: %s\n", ticker, volume, VerdictUndetermined)
				}
			} else {
				if timeStr != "" {
					fmt.Printf("%s Time: %s Verdict: %s\n", ticker, timeStr, VerdictUndetermined)
				} else {
					fmt.Printf("%s Verdict: %s\n", ticker, VerdictUndetermined)
				}
			}
		} else {
			if volume != "" {
				if timeStr != "" {
					fmt.Printf("%s Time: %s Volume: %s Verdict: %s\n", ticker, timeStr, volume, verdict)
				} else {
					fmt.Printf("%s Volume: %s Verdict: %s\n", ticker, volume, verdict)
				}
			} else {
				if timeStr != "" {
					fmt.Printf("%s Time: %s Verdict: %s\n", ticker, timeStr, verdict)
				} else {
					fmt.Printf("%s Verdict: %s\n", ticker, verdict)
				}
			}
		}
	}

	return newsItems
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

// FetchTickerNewsItem scrapes the news items for a given ticker from Finviz.
func FetchTickerNewsItem(ticker string) []NewsItem {
	url := "https://finviz.com/quote.ashx?t=" + ticker

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return []NewsItem{}
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return []NewsItem{}
	}

	var items []NewsItem

	selection := doc.Find("table#news-table tr")
	selection.Each(func(index int, s *goquery.Selection) {
		linkTag := s.Find("a")
		if linkTag.Length() == 0 {
			return
		}
		headline := linkTag.Text()
		href, _ := linkTag.Attr("href")
		timeOrDate := s.Find("td").First().Text()

		items = append(items, NewsItem{
			Headline: headline,
			Link:     href,
			Time:     timeOrDate,
		})

	
	})

	return items
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

