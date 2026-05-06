package command

import (
	"fmt"
	"os"

	"github.com/charmbracelet/log"

	"github.com/spf13/cobra"
)

var (
	// Command
	authCmd = &cobra.Command{
		Use:   "auth [org]",
		Short: "Authenticate to AWS",
		Long: `Establish an authenticated session with an AWS organization in a
specified region using a role.`,
		Args: cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			nargs := len(args)
			orgName := ""

			switch nargs {
			case 0:
				orgName, err = GetDefaultOrgName()

				if err != nil {
					log.Fatal(err)
				}
			default:
				orgName = args[0]
			}

			log.Debug("Loading org", "name", orgName)

			org, err := NewAwsOrgFromConfig(orgName)

			if err != nil {
				log.Fatal(err)
			}

			log.Debug("Loading cache")

			err = org.LoadFromCache()

			if err != nil {
				log.Fatal(err)
			}

			if org.Session.IsAccessTokenValid() {
				fmt.Printf("authenticated to %s\n", org.Data.Name)
				return
			}

			log.Debug("Initializing org", "org", orgName)

			err = org.Init()

			if err != nil {
				log.Fatal(err)
			}

			hostName, err := os.Hostname()

			if err != nil {
				log.Fatal(err)
			}

			log.Info("Acquired hostname", "hostname", hostName)

			err = org.StartSession(hostName)

			if err != nil {
				log.Fatal(err)
			}

			fmt.Println("verification url: ", org.Session.Data.VerificationUrl)
			fmt.Println("waiting for verification...")

			err = org.WaitForVerification(60)

			if err != nil {
				log.Fatal(err)
			}

			log.Debug("Saving session to cache", "org", orgName)

			err = org.Session.SaveToCache(orgName)

			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("authenticated to %s\n", org.Data.Name)
		},
	}
)

func init() {
	rootCmd.AddCommand(authCmd)
}
