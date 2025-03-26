package command

import (
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var (
	// Flags
	configPath string
	setDefault bool

	// Command
	saveCmd = &cobra.Command{
		Use:   "save",
		Short: "Saves the credentials to a file",
		Long:  `This command saves the cached credentials to the specified file path`,
		Args:  cobra.MinimumNArgs(1),
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
	rootCmd.AddCommand(saveCmd)

	saveCmd.Flags().StringVar(&configPath, "path", "", "Save the credentials to specified path")
	saveCmd.Flags().BoolVar(&setDefault, "default", false, "Sets the first account in selected list as the default")
}
