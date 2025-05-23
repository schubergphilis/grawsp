package command

import (
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var (
	// Command
	listCmd = &cobra.Command{
		Use:   "list [org]",
		Short: "List information about certain resources",
		Long: `The list command can be used to view the list of certain resources
such as accounts and organizations.`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			orgName := args[0]

			log.Debug("Organization", "orgName", orgName)
		},
	}
)

func init() {
	rootCmd.AddCommand(listCmd)
}
