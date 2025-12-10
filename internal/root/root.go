package root

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "assi",
	Short: "Assiarius CLI",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		return
	}
}
