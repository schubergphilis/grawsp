package command

import (
	"fmt"

	log "github.com/sirupsen/logrus"
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

			log.Debug("Loading list of accounts from cache")

			accounts, err := GetAccountsFromCache(orgName)

			if err != nil {
				log.Fatal(err)
			}

			for _, account := range accounts {
				fmt.Println(account.ID, account.Name, account.Email)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(listCmd)
}
