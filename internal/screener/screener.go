package screener

import (
	"fmt"
	"github.com/d3an/finviz/screener"
	"github.com/go-gota/gota/dataframe"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "screen [preset]",
		Short: "Run a single Finviz screener preset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunScreen(args[0])
		},
	}

	return cmd
}

func RunScreen(screen string) error {
	client := screener.New(nil)

	df, err := client.GetScreenerResults(screen)
	if err != nil {
		return fmt.Errorf("failed to fetch screener %q: %w", screen, err)
	}

	extractNewsSlice(df)
	return nil
}

func extractNewsSlice(df *dataframe.DataFrame) {
	records := df.Select(1).Records()
	for index, record := range records {
		ticker := record[0]
		url := "https://finviz.com/quote.ashx?t=" + ticker
		fmt.Println(index, ticker, url)
	}
}
