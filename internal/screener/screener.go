package screener

import (
	"fmt"
	"github.com/d3an/finviz/screener"
	"github.com/d3an/finviz/utils"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "screen [preset]",
		Short: "Run a single Finviz screener preset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runScreen(args[0])
		},
	}

	return cmd
}

func runScreen(screen string) error {
	client := screener.New(nil)

	df, err := client.GetScreenerResults("https://finviz.com/screener.ashx?v=111&s=ta_unusualvolume&f=exch_nasd,news_date_today,sh_relvol_o5&o=-volume")
	if err != nil {
		return fmt.Errorf("failed to fetch screener %q: %w", screen, err)
	}

	utils.PrintFullDataFrame(df)
	return nil
}
