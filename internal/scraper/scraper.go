package scraper

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"assiarius/internal/read"
	"assiarius/internal/tickerstats"

	"github.com/PuerkitoBio/goquery"
)

// This file contains ad-hoc scrapers for specific Finviz pages, used by the screener command.
func FetchTickerStatistics(ticker string) (*tickerstats.Stats, error) {
	doc, err := fetchFinvizQuoteDocument(ticker)
	if err != nil {
		return nil, err
	}
	stats := parseTickerStatisticsFromQuoteDocument(ticker, doc)
	if stats.Len() == 0 {
		return nil, fmt.Errorf("no statistics found for ticker %s", ticker)
	}

	return stats, nil
}

func GetTickerVolume(ticker string) (string, error) {
	return GetTickerValueAny(ticker, "Volume", "Avg Volume")
}

// GetTickerValue fetches a single Finviz snapshot statistic by its label/key.
//
// Examples of keys: "Volume", "Avg Volume", "Market Cap", "P/E".
// It logs and returns an error if the key is missing or the value is "-".
func GetTickerValue(ticker string, key string) (string, error) {
	return GetTickerValueAny(ticker, key)
}

// GetTickerValueAny tries multiple keys in order and returns the first
// non-missing value. It logs and returns an error if none are found.
func GetTickerValueAny(ticker string, keys ...string) (string, error) {
	ticker = strings.TrimSpace(strings.ToUpper(ticker))
	if ticker == "" {
		err := fmt.Errorf("ticker is empty")
		log.Printf("scraper: %v", err)
		return "", err
	}
	if len(keys) == 0 {
		err := fmt.Errorf("no keys provided for ticker %s", ticker)
		log.Printf("scraper: %v", err)
		return "", err
	}

	stats, err := FetchTickerStatistics(ticker)
	if err != nil {
		log.Printf("scraper: FetchTickerStatistics(%s) failed: %v", ticker, err)
		return "", err
	}

	cleanKeys := make([]string, 0, len(keys))
	for _, k := range keys {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		cleanKeys = append(cleanKeys, k)
		if v, ok := stats.Get(k); ok && !v.Missing() {
			out := strings.TrimSpace(v.Raw)
			if out != "" {
				return out, nil
			}
		}
	}

	err = fmt.Errorf("no value found for ticker %s (keys: %s)", ticker, strings.Join(cleanKeys, ", "))
	log.Printf("scraper: %v", err)
	return "", err
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