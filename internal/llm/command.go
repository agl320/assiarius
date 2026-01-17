package llm

import (
	"context"
	"fmt"
)

func ProcessText(ctx context.Context, text string, client Client) error {
	prompt := Prompt{
		Prompt:  "Determine verdict from the following text.",
		Message: text,
	}

	out, err := client.Process(ctx, prompt)
	if err != nil {
		return err
	}

	fmt.Println(out)
	return nil
}
