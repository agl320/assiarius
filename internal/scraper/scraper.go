package scraper

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gocolly/colly/v2"
)


func ReadRelativeVolume(ticker string) {
	fmt.Printf("Fetching relative volume for ticker: %s\n", ticker)
	url := "https://finviz.com/quote.ashx?t=" + ticker

	c := colly.NewCollector()

	c.OnHTML("table.snapshot-table2 tbody", func(e *colly.HTMLElement) {
		e.ForEach("tr", func(_ int, el *colly.HTMLElement) {
			key := ToKey(el.ChildText("td:nth-child(11)"))
			fmt.Printf("Ticker: %s, Key: %s\n", ticker, key)
		})})


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