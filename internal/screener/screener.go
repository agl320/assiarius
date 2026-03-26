package screener

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/d3an/finviz/screener"
	"github.com/go-gota/gota/dataframe"

	"assiarius/internal/scraper"
)

type NewsItem struct {
	Headline string
	Link     string
	Time     string
}

func RunScreen(screen string, includeNews bool) error {
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
		extractNewsSlice(df)
		return nil
	}

	printTickers(df)
	return nil
}

func printTickers(df *dataframe.DataFrame) {
	records := df.Records()
	for index, record := range records {
		if len(record) > 1 {
			ticker := cleanTicker(record[1])
			if ticker == "" {
				continue
			}
			fmt.Println(index, ticker)
		}
	}
}

func extractNewsSlice(df *dataframe.DataFrame) {
	records := df.Records()

	for index, record := range records {
		if len(record) > 0 {
			ticker := cleanTicker(record[1])
			if ticker == "" {
				continue
			}

			fmt.Println(index, ticker)
			GetNewsForTicker(ticker)
		}
	}
	
}

func cleanTicker(s string) string {
	s = strings.ToUpper(s)

	re := regexp.MustCompile(`[^A-Z0-9]`)
	s = re.ReplaceAllString(s, "")

	return s
}

func GetNewsForTicker(ticker string) []NewsItem {
	fmt.Println("News results:")

	newsItems := FetchTickerNewsItem(ticker)

	for i := 0; i < len(newsItems) && i < 1; i++ {
		item := newsItems[i]
		if strings.Contains(item.Time, "Today") {
			fmt.Printf("%d: %s - %s\n", i, item.Headline, item.Link)
		}

		text, links := scraper.ExtractNewsFromLink(item.Link)
		if text != "" {
			fmt.Println(text)
		}
		if len(links) > 0 {
			fmt.Printf("Found %d links in article\n", len(links))
		}
	}

	return newsItems
}

// FetchTickerNewsItem scrapes the news items for a given ticker from Finviz.
func FetchTickerNewsItem(ticker string) []NewsItem {
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
	defer resp.Body.Close()
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
