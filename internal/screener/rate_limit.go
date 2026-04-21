package screener

import (
	"os"
	"strings"
	"time"
)

// GetQueueMinInterval reads GEMINI_MIN_INTERVAL (duration string) or returns the default.
func GetQueueMinInterval() time.Duration {
	// Conservative default. You can lower it if your API limits allow.
	defaultInterval := 2 * time.Second

	raw := strings.TrimSpace(os.Getenv("GEMINI_MIN_INTERVAL"))
	if raw == "" {
		return defaultInterval
	}

	parsed, err := time.ParseDuration(raw)
	if err != nil {
		return defaultInterval
	}
	return parsed
}
