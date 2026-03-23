package scraper

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
)

// This file contains ad-hoc scrapers for specific Finviz pages, used by the screener command.
func ReadTickerStatistics(ticker string) {
	fmt.Printf("Fetching relative volume for ticker: %s\n", ticker)
	url := "https://finviz.com/quote.ashx?t=" + ticker

	c := colly.NewCollector()

	c.OnHTML("table.snapshot-table2 tbody", func(bodyEl *colly.HTMLElement) {
		bodyEl.ForEach("tr", func(_ int, rowEl *colly.HTMLElement) {

			var label string
			var value string

			rowEl.ForEach("td", func(_ int, cellEl *colly.HTMLElement) {
				classAttr := cellEl.Attr("class")

				if strings.Contains(classAttr, "cursor-pointer") {
					label = strings.TrimSpace(cellEl.Text)
					return
				}

				spanText := strings.TrimSpace(cellEl.ChildText("b span"))
				if spanText != "" {
					value = spanText
				}
			})

			if label != "" {
				fmt.Printf("Key: %-20s Value: %s\n", label, value)
			}
		})
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Error has occured:", err)
	})

	c.Visit(url)
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
	fmt.Println("Extracting news from link:", link)
	url := strings.TrimSpace(link)
	if url == "" {
		return "", []string{}
	}
	if (strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")) {
		fmt.Println("External link detected, skipping news extraction.")
		return "", []string{}
	}

	return extractNewsFromFinvizLink(url)
}

func extractNewsFromFinvizLink(url string) (string, []string) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		fmt.Println("error during page retrieval")
		return "", []string{}
	}

	defer resp.Body.Close()
	
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Println("error during page reading")
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