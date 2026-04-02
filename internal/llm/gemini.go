package llm

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/genai"
)

type GeminiClient struct {
	client *genai.Client
	model  string
	timeout time.Duration
}

func NewGeminiClient(ctx context.Context, cfg Config) (*GeminiClient, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: cfg.GeminiAPIKey})
	if err != nil {
		return nil, err
	}

	model := cfg.Model
	if model == "" {
		model = "models/gemini-3-flash-preview"
	}

	return &GeminiClient{
		client: client,
		model:  model,
		timeout: cfg.Timeout,
	}, nil
}

// Process sends a prompt to the Gemini model and returns the generated response.
func (g *GeminiClient) Process(ctx context.Context, p Prompt) (string, error) {
	if g.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, g.timeout)
		defer cancel()
	}

	result, err := g.client.Models.GenerateContent(
		ctx,
		g.model,
		genai.Text(
			fmt.Sprintf("%s\n\n%s", p.Prompt, p.Message),
		),
		nil,
	)
	if err != nil {
		return "", err
	}

	return result.Text(), nil
}
