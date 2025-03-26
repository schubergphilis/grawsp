package command

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	log "github.com/sirupsen/logrus"

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

				org, err := NewOrgFromConfig(orgName)

				if err != nil {
					log.Fatal(err)
				}

				log.Debug("AWS Region: ", org.Region)

				awsConfig, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(org.Region))

				if err != nil {
					log.Fatal(err)
				}

				log.Info("Establishing session with AWS organization")

				orgSession, err := NewOrgSessionFromCache(orgName)

				if err != nil {
					log.Fatal(err)
				}

				if !orgSession.IsAccessTokenValid() {
					log.Error("You are not authenticated!")
					os.Exit(1)
				}

				log.Debug("Initializing org session")

				if err := orgSession.Init(awsConfig); err != nil {
					log.Fatal(err)
				}

				log.Debug("Fetching account list")

				accounts, err := orgSession.ListAccounts()

				if err != nil {
					log.Fatal(err)
				}

				log.Debug("Caching accounts")

				err = WriteAccountsToCache(orgName, accounts)

				if err != nil {
					log.Fatal(err)
				}

				log.Infof("%d accounts synchronized!", len(accounts))
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(syncCmd)
}
