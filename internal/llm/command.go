package llm

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "process [text]",
		Short: "Determine verdict from given text using LLM",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return ProcessText(cmd.Context(), args[0])
		},
	}

	return cmd
}

func ProcessText(ctx context.Context, text string) error {
	client, err := NewGeminiClient(ctx)
	if err != nil {
		return err
	}

	prompt := Prompt{
		Prompt: "Determine verdict from the following text.",
		Message:   text,
	}

	out, err := client.Process(ctx, prompt)
	if err != nil {
		return err
	}

	fmt.Println(out)
	return nil
}