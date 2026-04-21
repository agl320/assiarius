package screener

import (
	"strings"
)

func FormatWebhookResult(res ScreenResult) string {
	lines := make([]string, 0, 6)

	ticker := strings.TrimSpace(strings.ToUpper(res.Ticker))
	ticker = strings.TrimPrefix(ticker, "$")
	if ticker == "" {
		return ""
	}

	if h := strings.TrimSpace(strings.Join(strings.Fields(res.Latest.Headline), " ")); h != "" {
		lines = append(lines, "## "+h)
		lines = append(lines, "**$"+ticker+"**")
	} else {
		lines = append(lines, "**$"+ticker+"**")
	}
	verdict := strings.TrimSpace(res.Verdict)
	volume := strings.TrimSpace(res.Volume)
	metaParts := make([]string, 0, 4)
	if verdict != "" {
		metaParts = append(metaParts, "Verdict:", verdict)
	}
	if volume != "" {
		metaParts = append(metaParts, "Volume:", volume)
	}
	if len(metaParts) > 0 {
		lines = append(lines, strings.Join(metaParts, " "))
	}

	// Remaining lines: Time + Link
	if t := strings.TrimSpace(strings.Join(strings.Fields(res.Latest.Time), " ")); t != "" {
		lines = append(lines, "Time: "+t)
	}
	if link := strings.TrimSpace(res.Latest.Link); link != "" {
		lines = append(lines, "Link: "+link)
	}

	return strings.Join(lines, "\n")
}

func FormatWebhookTicker(t ScreenTicker) string {
	ticker := strings.TrimSpace(strings.ToUpper(t.Ticker))
	ticker = strings.TrimPrefix(ticker, "$")
	if ticker == "" {
		return ""
	}

	lines := make([]string, 0, 2)
	lines = append(lines, "**$"+ticker+"**")
	if v := strings.TrimSpace(t.Volume); v != "" {
		lines = append(lines, "Volume: "+v)
	}
	return strings.Join(lines, "\n")
}
