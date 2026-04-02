package scraper

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"assiarius/internal/read"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
)

// This file contains ad-hoc scrapers for specific Finviz pages, used by the screener command.
func FetchTickerStatistics(ticker string) (map[string]string, error) {
	url := "https://finviz.com/quote.ashx?t=" + ticker

	stats := map[string]string{}
	var callbackErr error

	c := colly.NewCollector()
	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	c.SetRequestTimeout(10 * time.Second)

	c.OnHTML("table.snapshot-table2 tbody", func(bodyEl *colly.HTMLElement) {
		bodyEl.ForEach("tr", func(_ int, rowEl *colly.HTMLElement) {
			cells := rowEl.DOM.Find("td")
			for i := 0; i+1 < cells.Length(); i += 2 {
				label := strings.TrimSpace(cells.Eq(i).Text())
				value := strings.TrimSpace(cells.Eq(i + 1).Text())
				if label == "" {
					continue
				}
				stats[label] = value
			}
		})
	})

	c.OnError(func(r *colly.Response, err error) {
		callbackErr = err
	})

	if err := c.Visit(url); err != nil {
		return nil, err
	}
	if callbackErr != nil {
		return nil, callbackErr
	}
	if len(stats) == 0 {
		return nil, fmt.Errorf("no statistics found for ticker %s", ticker)
	}

	return stats, nil
}

func GetTickerVolume(ticker string) (string, error) {
	stats, err := FetchTickerStatistics(ticker)
	if err != nil {
		return "", err
	}
	if v, ok := stats["Volume"]; ok && strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v), nil
	}
	if v, ok := stats["Avg Volume"]; ok && strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v), nil
	}
	return "", fmt.Errorf("volume not found for ticker %s", ticker)
}

func ToKey(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)

	re := regexp.MustCompile(`[^a-z0-9]+`)
	s = re.ReplaceAllString(s, "_")

	s = strings.Trim(s, "_")

	return s
}


func ExtractNewsFromLink(link string) (string, []string) {
	url := strings.TrimSpace(link)
	if url == "" {
		return "", []string{}
	}
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return extractNewsFromExternalLink(url)
	}
	url = "https://finviz.com" + url
	return extractNewsFromFinvizLink(url)
}

func extractNewsFromExternalLink(url string) (string, []string) {
	text, err := read.ReadNewsTextFromLink(url)
	if err != nil {
		return "", []string{}
	}
	return text, []string{}
}

func extractNewsFromFinvizLink(url string) (string, []string) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return "", []string{}
	}

	defer resp.Body.Close()
	
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", []string{}
	}

	selection := doc.Find("div.text-justify")
	text := selection.Text()
	clean := strings.Join(strings.Fields(text), " ")
	
	var links []string
	selection.Find("a").Each(func(i int, s *goquery.Selection) {
		if href, ok := s.Attr("href"); ok {
			links = append(links, href)
		}
	})
	return clean, links
}