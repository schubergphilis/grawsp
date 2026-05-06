package command

import (
	"fmt"
	"os"
	"runtime"
	"sync"

	"github.com/charmbracelet/log"
	"github.com/schubergphilis/grawsp/internal/awsservice"

	"github.com/spf13/cobra"
)

var (
	// Globals
	workersCount int

	// Flags
	withRoles bool

	// Command
	syncCmd = &cobra.Command{
		Use:   "sync [org]",
		Short: "Download list of accounts from Organization",
		Long: `This command downloads the available accounts and roles to be used
    and caches them locally.`,
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

			if !org.Session.IsAccessTokenValid() {
				log.Error("you are not authenticated to the org", "org", orgName)
				os.Exit(1)
			}

			log.Info("You are authenticated to the org", "org", orgName)
			log.Debug("Initializing org", "org", orgName)

			err = org.Init()

			if err != nil {
				log.Fatal(err)
			}

			log.Debug("Synchronizing accounts")

			err = org.SyncAccounts()

			if err != nil {
				log.Fatal(err)
			}

			if withRoles {
				var wg sync.WaitGroup

				jobs := make(chan *awsservice.AwsAccount)

				log.Debug("Number of workers", "workers", workersCount)

				for i := 1; i <= workersCount; i++ {
					wg.Go(func() {
						for account := range jobs {
							log.Debug("Syncing account roles", "account", account.Data.Name, "worker", i)

							err := account.SyncRoles(org.Session.Data.AccessToken)

							if err != nil {
								log.Error(err)
							}
						}
					})
				}

				wg.Go(func() {
					for account := range org.Accounts.All() {
						jobs <- account
					}

					close(jobs)
				})

				wg.Wait()
			}

			log.Debug("Saving org to cache", "org", orgName)

			err = org.SaveToCache()

			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("%s synchronised\n", org.Data.Name)
		},
	}
)

func init() {
	workersCount = runtime.NumCPU()

	syncCmd.Flags().BoolVar(&withRoles, "roles", false, "sync the roles of the accounts")

	rootCmd.AddCommand(syncCmd)
}
