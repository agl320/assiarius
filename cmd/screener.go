package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"assiarius/internal/screener"
	"assiarius/internal/webhook"

	"github.com/spf13/cobra"
)

func screenerCommand() *cobra.Command {
	var includeNews bool
	var sendWebhook bool

	cmd := &cobra.Command{
		Use:   "screen [preset]",
		Short: "Run a single Finviz screener preset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Printf("screen: starting preset=%s includeNews=%v", args[0], includeNews)
			run, err := screener.RunScreen(cmd.Context(), args[0], includeNews, app.LLM)
			if err != nil {
				return err
			}
			log.Printf("screen: tickers=%d results=%d", len(run.Tickers), len(run.Results))
			if len(run.Tickers) == 0 {
				fmt.Printf("No results found for screener %q\n", args[0])
				return nil
			}

			if includeNews {
				for _, res := range run.Results {
					fmt.Println(screener.FormatScreenResult(res))
					if sendWebhook {
						payload := screener.FormatWebhookResult(res)
						if payload != "" {
							if err := webhook.NotifyDiscord(cmd.Context(), payload); err != nil {
								fmt.Fprintln(os.Stderr, err)
								log.Printf("screen: webhook error: %v", err)
							}
						}
					}
				}
				return nil
			}

			for i, t := range run.Tickers {
				ticker := strings.TrimSpace(t.Ticker)
				if ticker == "" {
					continue
				}
				if v := strings.TrimSpace(t.Volume); v != "" {
					fmt.Printf("%d %s Volume: %s\n", i+1, ticker, v)
				} else {
					fmt.Printf("%d %s\n", i+1, ticker)
				}
				if sendWebhook {
					payload := screener.FormatWebhookTicker(t)
					if payload != "" {
						if err := webhook.NotifyDiscord(cmd.Context(), payload); err != nil {
							fmt.Fprintln(os.Stderr, err)
							log.Printf("screen: webhook error: %v", err)
						}
					}
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&includeNews, "news", false, "Fetch per-ticker news for screener results")
	cmd.Flags().BoolVar(&sendWebhook, "webhook", false, "Send output to Discord via WEBHOOK_DISCORD_URL")

	return cmd
}

func fetchTickerNewsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "news [ticker]",
		Short: "Fetch news for a ticker",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := screener.GetNewsForTicker(cmd.Context(), args[0], app.LLM)
			if err != nil {
				return err
			}
			fmt.Println(screener.FormatScreenResult(res))
			return nil
		},
	}

	return cmd
}