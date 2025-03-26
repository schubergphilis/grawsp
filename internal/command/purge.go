package command

import (
	"github.com/spf13/cobra"
)

var (
	// Command
	purgeCmd = &cobra.Command{
		Use:   "purge",
		Short: "Removes cached data",
		Long:  `Remove credentials and other metadata from the caching store.`,
		Run: func(cmd *cobra.Command, args []string) {
		},
	}
)

func init() {
	rootCmd.AddCommand(purgeCmd)
}
