package llm

import "time"

// Config holds configuration for LLM interactions, such as API keys and model settings.
type Config struct {
	GeminiAPIKey string
	Model        string
	Timeout      time.Duration
}