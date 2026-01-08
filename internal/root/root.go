package root

import (
	"assiarius/internal/debug"
	"assiarius/internal/poll"
	"assiarius/internal/read"
	"assiarius/internal/screener"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "assi",
	Short: "Assiarius CLI",
}

func init() {
	rootCmd.AddCommand(screener.Command())
	rootCmd.AddCommand(poll.Command())
	rootCmd.AddCommand(debug.Command())
	rootCmd.AddCommand(read.Command())
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
