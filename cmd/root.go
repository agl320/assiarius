package cmd

import (
	"assiarius/internal/llm"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

type App struct {
	LLM llm.Client
}

var app App

var rootCmd = &cobra.Command{
	Use:   "assi",
	Short: "Assiarius CLI",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		_ = godotenv.Load()

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		cfg := llm.Config{
			GeminiAPIKey: os.Getenv("GEMINI_API_KEY"),
		}

		client, err := llm.NewGeminiClient(ctx, cfg)
		if err != nil {
			return err
		}

		app = App{
			LLM: client,
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(screenerCommand())
	rootCmd.AddCommand(pollCommand())
	rootCmd.AddCommand(readCommand())
	rootCmd.AddCommand(llmCommand(&app))
}

	func Execute() {
		err := rootCmd.Execute()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
