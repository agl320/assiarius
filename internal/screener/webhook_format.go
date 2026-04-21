package screener

import (
	"fmt"
	"strings"
)

// FormatWebhookResult formats a Discord-friendly message for a screen/poll result.
// Includes news fields (time/headline/link) when present.
func FormatWebhookResult(res ScreenResult) string {
	lines := make([]string, 0, 6)

	ticker := strings.TrimSpace(strings.ToUpper(res.Ticker))
	if ticker == "" {
		ticker = "(UNKNOWN)"
	}

	// First line: compact summary.
	summaryParts := []string{ticker}
	if v := strings.TrimSpace(res.Verdict); v != "" {
		summaryParts = append(summaryParts, "Verdict:", v)
	}
	if vol := strings.TrimSpace(res.Volume); vol != "" {
		summaryParts = append(summaryParts, "Volume:", vol)
	}
	lines = append(lines, strings.Join(summaryParts, " "))

	if t := strings.TrimSpace(strings.Join(strings.Fields(res.Latest.Time), " ")); t != "" {
		lines = append(lines, "Time: "+t)
	}
	if h := strings.TrimSpace(strings.Join(strings.Fields(res.Latest.Headline), " ")); h != "" {
		lines = append(lines, "Headline: "+h)
	}
	if link := strings.TrimSpace(res.Latest.Link); link != "" {
		lines = append(lines, "Link: "+link)
	}

	return strings.Join(lines, "\n")
}

// FormatWebhookTicker formats a Discord-friendly message for a screener ticker
// when no news analysis is being performed.
func FormatWebhookTicker(t ScreenTicker) string {
	ticker := strings.TrimSpace(strings.ToUpper(t.Ticker))
	if ticker == "" {
		return ""
	}
	if v := strings.TrimSpace(t.Volume); v != "" {
		return fmt.Sprintf("%s Volume: %s", ticker, v)
	}
	return ticker
}
