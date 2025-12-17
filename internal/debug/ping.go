package debug

import (
	"fmt"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ping",
		Short: "Ping Assiarius",
		RunE: func(cmd *cobra.Command, args []string) error {
			ping()
			return nil
		},
	}

	return cmd
}

func ping() {
	fmt.Println("Assiarius ping received.")
}
