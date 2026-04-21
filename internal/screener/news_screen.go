package screener

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

// ValidateFinvizScreenerURL ensures the input is a full Finviz screener URL.
// Polling relies on being able to edit the query string.
//
// Note: Polling can construct the news window filter dynamically; the base URL
// does not need to include a news_date_prev* filter.
func ValidateFinvizScreenerURL(raw string) (*url.URL, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("screener url is empty")
	}

	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid screener url: %w", err)
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return nil, fmt.Errorf("screener url must include scheme (https://)")
	}
	if u.Host == "" {
		return nil, fmt.Errorf("screener url must include host (finviz.com)")
	}
	if !strings.Contains(strings.ToLower(u.Host), "finviz.com") {
		return nil, fmt.Errorf("screener url must be a finviz.com URL")
	}
	if !strings.HasSuffix(strings.ToLower(u.Path), "/screener.ashx") {
		return nil, fmt.Errorf("screener url must point to /screener.ashx")
	}

	return u, nil
}

func isNewsDateFilter(filter string) bool {
	filter = strings.TrimSpace(filter)
	return strings.HasPrefix(filter, "news_date_prevminutes") || strings.HasPrefix(filter, "news_date_prevhours")
}

// BuildNewsScreenerURL returns a new screener URL that includes a news time window
// filter (news_date_prevminutesX or news_date_prevhoursX).
//
// The returned URL is suitable for polling "past N minutes" where N is derived
// from window.
func BuildNewsScreenerURL(base string, window time.Duration) (string, error) {
	u, err := ValidateFinvizScreenerURL(base)
	if err != nil {
		return "", err
	}
	if window <= 0 {
		return "", fmt.Errorf("window must be > 0")
	}

	var newsFilter string
	if window >= time.Hour && window%time.Hour == 0 {
		hours := int(window / time.Hour)
		newsFilter = fmt.Sprintf("news_date_prevhours%d", hours)
	} else if window >= time.Minute && window%time.Minute == 0 {
		minutes := int(window / time.Minute)
		newsFilter = fmt.Sprintf("news_date_prevminutes%d", minutes)
	} else {
		return "", fmt.Errorf("window must be a whole number of minutes or hours (got %s)", window)
	}

	q := u.Query()
	filtersRaw := strings.TrimSpace(q.Get("f"))
	var filters []string
	if filtersRaw != "" {
		filters = strings.Split(filtersRaw, ",")
	}

	out := make([]string, 0, len(filters)+1)
	out = append(out, newsFilter)
	for _, f := range filters {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		if isNewsDateFilter(f) {
			continue
		}
		out = append(out, f)
	}

	q.Set("f", strings.Join(out, ","))
	u.RawQuery = q.Encode()
	return u.String(), nil
}
