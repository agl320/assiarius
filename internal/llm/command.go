package llm

import (
	"context"
	"errors"
	"fmt"
)

// This file contains logic for interacting with LLMs, such as formatting prompts and processing responses.
func ProcessText(ctx context.Context, text string, client Client) error {
	prompt := Prompt{
		Prompt:  "Determine verdict from the following text.",
		Message: text,
	}

	out, err := client.Process(ctx, prompt)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil
		}
		return err
	}

	fmt.Println(out)
	return nil
}
