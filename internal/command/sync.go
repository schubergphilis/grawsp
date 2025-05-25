package command

import (
	"fmt"

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
				log.Debug("Loading org", "name", orgName)

				org, err := NewAwsOrgFromConfig(orgName)

				if err != nil {
					log.Fatal(err)
				}

				if !org.Session.IsAccessTokenValid() {
					log.Fatal("You are not authenticated to org", "org", orgName)
				}

				log.Debug("You are authenticated to org", "org", orgName)

				err = org.SyncAccounts()

				if err != nil {
					log.Fatal(err)
				}

				fmt.Printf("Accounts of %s synchronised\n", org.Data.Name)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(syncCmd)
}
