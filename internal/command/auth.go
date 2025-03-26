package command

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc/types"
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

				if orgSession.IsAccessTokenValid() {
					fmt.Println("Authentication successful!")
					return
				}

				log.Debug("Initializing org session")

				if err := orgSession.Init(awsConfig); err != nil {
					log.Fatal(err)
				}

				hostName, err := os.Hostname()

				if err != nil {
					log.Fatal(err)
				}

				log.Debug("Acquired hostname: ", hostName)

				if err = orgSession.Authenticate(hostName, org.StartUrl); err != nil {
					log.Fatal(err)
				}

				fmt.Println("Verification URL: ", orgSession.VerificationUrl)
				fmt.Println("Waiting for authorization...")

				for {
					err = orgSession.CreateAccessToken()

					if err == nil {
						break
					}

					var bne *types.AuthorizationPendingException

					if errors.As(err, &bne) {
						time.Sleep(3 * time.Second)
					} else {
						log.Fatal(err)
					}
				}

				log.Debug("Caching org session")

				err = WriteOrgSessionToCache(orgName, orgSession)

				if err != nil {
					log.Fatal(err)
				}

				fmt.Println("Authentication successful!")
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(authCmd)
}
