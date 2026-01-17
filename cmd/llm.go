package cmd

import (
	"assiarius/internal/llm"

	"github.com/spf13/cobra"
)

func llmCommand(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "process [text]",
		Short: "Determine verdict from given text using LLM",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return llm.ProcessText(cmd.Context(), args[0], app.LLM)
		},
	}

	return cmd
}