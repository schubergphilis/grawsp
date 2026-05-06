package command

import (
	"github.com/charmbracelet/log"
	"github.com/schubergphilis/grawsp/internal/cacheservice"
	"github.com/spf13/cobra"
)

var (
	// Command
	purgeCmd = &cobra.Command{
		Use:   "purge [org]",
		Short: "Removes cached data",
		Long:  `Remove credentials and other metadata from the caching store.`,
		Run: func(cmd *cobra.Command, args []string) {
			var err error

			orgName := args[0]

			if orgName == "" {
				orgName, err = GetDefaultOrgName()

				if err != nil {
					log.Fatal(err)
				}
			}

			err = cacheservice.Purge(orgName)

			if err != nil {
				log.Fatal(err)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(purgeCmd)
}
