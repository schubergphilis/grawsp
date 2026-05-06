package command

import (
	"fmt"
	"os"

	"github.com/charmbracelet/log"

	"github.com/spf13/cobra"
)

var (
	// Command
	accessCmd = &cobra.Command{
		Use:   "access [org] [account] [role]",
		Short: "Get credentials for AWS accounts",
		Long: `The access command can be used to retrieve temporary credentials
to access the selected AWS accounts.`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var err error

			orgName := ""
			accountName := ""
			roleName := ""
			nargs := len(args)

			switch nargs {
			case 1:
				accountName = args[0]
			case 2:
				accountName = args[0]
				roleName = args[1]
			case 3:
				orgName = args[0]
				accountName = args[1]
				roleName = args[2]
			default:
				log.Error("Invalid number of arguments", "nargs", nargs)
				os.Exit(1)
			}

			if orgName == "" {
				orgName, err = GetDefaultOrgName()

				if err != nil {
					log.Fatal(err)
				}
			}

			org, err := NewAwsOrgFromConfig(orgName)

			if err != nil {
				log.Fatal(err)
			}

			log.Debug("Loading cache")

			err = org.LoadFromCache()

			if err != nil {
				log.Fatal(err)
			}

			log.Debug("Initializing org", "org", orgName)

			err = org.Init()

			if err != nil {
				log.Fatal(err)
			}

			if !org.Session.IsAccessTokenValid() {
				log.Fatal("You are not authenticated to org", "org", org.Data.Name)
			}

			if roleName == "" {
				if val, ok := org.Data.Roles["default"]; ok {
					roleName = val
				} else {
					log.Fatal("Default role not found")
				}
			}

			log.Debug("Using org...", "org", orgName)
			log.Debug("Using role...", "role", roleName)

			account := org.Accounts.FindByName(accountName)

			if account == nil {
				log.Error("Account not found", "account", accountName)
				os.Exit(1)
			}

			log.Debug("Using account...", "account", account.Data.Name)

			role := account.Roles.FindByName(roleName)

			if role == nil {
				log.Debug("Syncing account roles", "account", account.Data.Name)

				err = account.SyncRoles(org.Session.Data.AccessToken)

				if err != nil {
					log.Fatal(err)
				}

				role = account.Roles.FindByName(roleName)

				if role == nil {
					log.Error("Role not found", "role", roleName)
					os.Exit(1)
				}
			}

			if !role.IsCredentialValid() {
				log.Debug("Getting credentials", "role", roleName)

				err = role.GetCredential(org.Session.Data.AccessToken, account.Data.ID)

				if err != nil {
					log.Fatal(err)
				}
			} else {
				log.Debug("Credentials are still valid", "role", roleName)
			}

			log.Debug("Saving account to cache", "org", orgName)

			err = account.SaveToCache(orgName)

			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("access granted\n")

			consoleUrl, err := role.GetConsoleURL(org.Data.Region)

			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(consoleUrl)
		},
	}
)

func init() {
	rootCmd.AddCommand(accessCmd)
}
