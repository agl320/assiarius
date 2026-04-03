package scraper

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"assiarius/internal/tickerstats"

	"github.com/PuerkitoBio/goquery"
)

const finvizUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

// NewsItem is a single entry from Finviz's quote page news table.
//
// NOTE: the Time field may contain either time-only or date+time depending on
// where it appears in the table.
//
type NewsItem struct {
	Headline string
	Link     string
	Time     string
}

func finvizQuoteURL(ticker string) string {
	ticker = strings.TrimSpace(strings.ToUpper(ticker))
	return "https://finviz.com/quote.ashx?t=" + ticker
}

func fetchFinvizQuoteDocument(ticker string) (*goquery.Document, error) {
	ticker = strings.TrimSpace(strings.ToUpper(ticker))
	if ticker == "" {
		return nil, fmt.Errorf("ticker is empty")
	}

	url := finvizQuoteURL(ticker)
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", finvizUserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("finviz quote request failed for %s: status=%d", ticker, resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func parseTickerStatisticsFromQuoteDocument(ticker string, doc *goquery.Document) *tickerstats.Stats {
	stats := tickerstats.New(ticker)
	if doc == nil {
		return stats
	}

	doc.Find("table.snapshot-table2 tr").Each(func(_ int, row *goquery.Selection) {
		cells := row.Find("td")
		for i := 0; i+1 < cells.Length(); i += 2 {
			label := strings.TrimSpace(cells.Eq(i).Text())
			value := strings.TrimSpace(cells.Eq(i + 1).Text())
			if label == "" {
				continue
			}
			stats.Set(label, value)
		}
	})

	return stats
}

func parseTickerNewsFromQuoteDocument(doc *goquery.Document) []NewsItem {
	if doc == nil {
		return []NewsItem{}
	}

	items := make([]NewsItem, 0, 16)
	doc.Find("table#news-table tr").Each(func(_ int, row *goquery.Selection) {
		linkTag := row.Find("a")
		if linkTag.Length() == 0 {
			return
		}
		headline := strings.TrimSpace(linkTag.Text())
		href, _ := linkTag.Attr("href")	// href may be missing; keep as empty string
		timeOrDate := strings.TrimSpace(row.Find("td").First().Text())

		items = append(items, NewsItem{Headline: headline, Link: href, Time: timeOrDate})
	})

	return items
}

// FetchTickerQuote fetches the Finviz quote page once and extracts both
// statistics and news items.
func FetchTickerQuote(ticker string) (*tickerstats.Stats, []NewsItem, error) {
	doc, err := fetchFinvizQuoteDocument(ticker)
	if err != nil {
		return nil, nil, err
	}
	stats := parseTickerStatisticsFromQuoteDocument(ticker, doc)
	news := parseTickerNewsFromQuoteDocument(doc)
	return stats, news, nil
}

// FetchTickerNewsItem scrapes the news items for a given ticker from Finviz.
func FetchTickerNewsItem(ticker string) []NewsItem {
	doc, err := fetchFinvizQuoteDocument(ticker)
	if err != nil {
		return []NewsItem{}
	}
	return parseTickerNewsFromQuoteDocument(doc)
}
