package llm

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

	type GeminiClient struct {
		client *genai.Client
		model  string
	}

	func NewGeminiClient(ctx context.Context, cfg Config) (*GeminiClient, error) {
		client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: cfg.GeminiAPIKey})
		if err != nil {
			return nil, err
		}

		return &GeminiClient{
			client: client,
			model:  "models/gemini-3-flash-preview",
		}, nil
	}

	func (g *GeminiClient) Process(ctx context.Context, p Prompt) (string, error) {
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
