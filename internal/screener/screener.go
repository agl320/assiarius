package screener

import (
	"assiarius/internal/scraper"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/d3an/finviz/screener"
	"github.com/go-gota/gota/dataframe"
)

type NewsItem struct {
	Headline string
	Link     string
	Time     string
}

func RunScreen(screen string) error {
	client := screener.New(nil)

	df, err := client.GetScreenerResults(screen)
	if err != nil {
		return fmt.Errorf("failed to fetch screener %q: %w", screen, err)
	}

	extractNewsSlice(df)
	return nil
}

func extractNewsSlice(df *dataframe.DataFrame) {
	colNames := df.Names()
	records := df.Records()

	fixedRecords := append([][]string{colNames}, records...)

	for index, record := range fixedRecords {
		if len(record) > 0 {
			ticker := cleanTicker(record[1])
			if ticker == "" {
				continue
			}
			scraper.ReadRelativeVolume(ticker)
			newsSlice := fetchTickerNewsItem(ticker)
			fmt.Println(index, ticker, len(newsSlice))
		}
	}
}

func cleanTicker(s string) string {
	s = strings.ToUpper(s)

	re := regexp.MustCompile(`[^A-Z0-9]`)
	s = re.ReplaceAllString(s, "")

	return s
}

func fetchTickerNewsItem(ticker string) []NewsItem {
	url := "https://finviz.com/quote.ashx?t=" + ticker
	fmt.Println(url)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		fmt.Println("error during page retrieval")
		return []NewsItem{}
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Println("error during page reading")
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
