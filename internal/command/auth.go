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
		Use:   "auth [org...]",
		Short: "Authenticate to AWS",
		Long: `Establish an authenticated session with an AWS organization in a
specified region using a role.`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			for _, orgName := range args {
				log.Debug("Loading org", "name", orgName)

				org, err := NewAwsOrgFromConfig(orgName)

				if err != nil {
					log.Fatal(err)
				}

				if org.Session.IsAccessTokenValid() {
					fmt.Printf("Authenticated to %s\n", org.Data.Name)
					continue
				}

				hostName, err := os.Hostname()

				if err != nil {
					log.Fatal(err)
				}

				log.Debug("Acquired hostname", "hostname", hostName)

				err = org.StartSession(hostName)

				if err != nil {
					log.Fatal(err)
				}

				fmt.Println("Verification URL: ", org.Session.Data.VerificationUrl)
				fmt.Println("Waiting for verification...")

				err = org.WaitForVerification(60.0)

				if err != nil {
					log.Fatal(err)
				}

				fmt.Printf("Authenticated to %s\n", org.Data.Name)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(authCmd)
}
