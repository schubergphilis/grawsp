package command

import (
	"github.com/charmbracelet/log"

	"github.com/spf13/cobra"
)

var (
	// Command
	accessCmd = &cobra.Command{
		Use:   "access [selector] [role]",
		Short: "Get credentials for AWS accounts",
		Long: `The access command can be used to retrieve temporary credentials
to access the selected AWS accounts.`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			selector := args[0]
			role := ""

			if len(args) > 1 {
				role = args[1]
			}

			log.Debug("Selector: ", selector)
			log.Debug("Role: ", role)
		},
	}
)

func init() {
	rootCmd.AddCommand(accessCmd)
}
