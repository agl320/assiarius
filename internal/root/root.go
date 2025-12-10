package root

import (
	"assiarius/internal/screener"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "assi",
	Short: "Assiarius CLI",
}

func init() {
	rootCmd.AddCommand(screener.Command())
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
