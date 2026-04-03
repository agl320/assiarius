package scraper

import (
	"fmt"
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