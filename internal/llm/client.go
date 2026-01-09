package llm

import "context"

type Client interface {
	Process(ctx context.Context, prompt Prompt) (string, error)
}