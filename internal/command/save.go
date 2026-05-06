package command

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"
)

var (
	// Flags
	configPath string
	setDefault bool

	// Command
	saveCmd = &cobra.Command{
		Use:   "save [org] [account] [role]",
		Short: "Saves the credentials to a file",
		Long:  `This command saves the cached credentials to the specified file path`,
		Args:  cobra.MinimumNArgs(1),
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

			log.Debug("Using org...", "org", orgName)
			log.Debug("Initializing org", "org", orgName)

			err = org.Init()

			if err != nil {
				log.Fatal(err)
			}

			account := org.Accounts.FindByName(accountName)

			if account == nil {
				log.Error("Account not found", "account", accountName)
				os.Exit(1)
			}

			log.Debug("Using account...", "account", account.Data.Name)

			if roleName == "" {
				if val, ok := org.Data.Roles["default"]; ok {
					roleName = val
				} else {
					log.Error("No default role is set for this org")
					os.Exit(1)
				}
			}

			log.Debug("Using role...", "role", roleName)

			role := account.Roles.FindByName(roleName)

			if role == nil {
				log.Error("Role not found", "role", roleName)
				os.Exit(1)
			}

			if !role.IsCredentialValid() {
				log.Error("No valid credentials were found", "account", accountName, "role", roleName)
				os.Exit(1)
			}

			if configPath == "" {
				homeDir, err := os.UserHomeDir()

				if err != nil {
					log.Fatal(err)
				}

				configPath = filepath.Join(homeDir, ".aws/credentials")
			}

			credentialsFile, err := ini.Load(configPath)

			if err != nil {
				fmt.Printf("Fail to read credentials file: %v", err)
				os.Exit(1)
			}

			var sectionName string

			if setDefault {
				sectionName = "default"
			} else {
				sectionName = fmt.Sprintf("%s-%s", accountName, roleName)
			}

			section, err := credentialsFile.NewSection(sectionName)

			if err != nil {
				log.Fatal(err)
			}

			section.Key("aws_access_key_id").SetValue(role.Credential.AccessKeyId)
			section.Key("aws_secret_access_key").SetValue(role.Credential.SecretAccessKey)
			section.Key("aws_session_token").SetValue(role.Credential.SessionToken)

			credentialsFile.SaveTo(configPath)
		},
	}
)

func init() {
	rootCmd.AddCommand(saveCmd)

	saveCmd.Flags().StringVar(&configPath, "path", "", "Save the credentials to specified path")
	saveCmd.Flags().BoolVar(&setDefault, "default", false, "Sets the account as the default")
}
