package cmd

import (
	"assiarius/internal/screener"

	"github.com/spf13/cobra"
)

func screenerCommand() *cobra.Command {
	var includeNews bool

	cmd := &cobra.Command{
		Use:   "screen [preset]",
		Short: "Run a single Finviz screener preset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return screener.RunScreen(cmd.Context(), args[0], includeNews, app.LLM)
		},
	}

	cmd.Flags().BoolVar(&includeNews, "news", false, "Fetch per-ticker news for screener results")

	return cmd
}

func fetchTickerNewsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "news [ticker]",
		Short: "Fetch news for a ticker",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			screener.GetNewsForTicker(cmd.Context(), args[0], app.LLM)
			return nil
		},
	}

	return cmd
}