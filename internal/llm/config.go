package llm

import "time"

type Config struct {
	GeminiAPIKey string
	Model        string
	Timeout      time.Duration
}