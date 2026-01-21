package scraper

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gocolly/colly/v2"
)

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
