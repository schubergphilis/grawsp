package command

import (
	"github.com/charmbracelet/log"

	"github.com/spf13/cobra"
)

var (
	// Command
	syncCmd = &cobra.Command{
		Use:   "sync [orgs...]",
		Short: "Download list of accounts from Organization",
		Long: `This command downloads the available accounts and roles to be used
    and caches them locally.`,
		Args: cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			for _, orgName := range args {
				log.Debug("Loading org from config: ", orgName)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(syncCmd)
}
