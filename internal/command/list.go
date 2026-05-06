package command

import (
	"fmt"
	"os"
	"regexp"
	"text/tabwriter"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var (
	// Flags
	filter     string
	printEmail bool
	printId    bool
	printRoles bool

	// Command
	listCmd = &cobra.Command{
		Use:   "list [org]",
		Short: "List information about certain resources",
		Long: `The list command can be used to view the list of certain resources
such as accounts and organizations.`,
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

			log.Debug("Compiling account filter")

			pattern, err := regexp.Compile(filter)

			if err != nil {
				log.Fatal(err)
			}

			writer := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
			defer writer.Flush()

			for account := range org.Accounts.All() {
				if !pattern.MatchString(account.Data.Name) {
					continue
				}

				line := account.Data.Name

				if printId {
					line = account.Data.ID + "\t" + account.Data.Name
				}

				if printRoles {
					line += "\t" + account.Roles.ToString()
				}

				if printEmail {
					line += "\t" + account.Data.Email
				}

				fmt.Fprintln(writer, line)
			}
		},
	}
)

func init() {
	listCmd.Flags().StringVar(&filter, "filter", "^.*$", "filter the accounts by their name")
	listCmd.Flags().BoolVar(&printId, "id", false, "print the ID of the account")
	listCmd.Flags().BoolVar(&printEmail, "email", false, "print the E-mail assigned to the account")
	listCmd.Flags().BoolVar(&printRoles, "roles", false, "print the Roles associated with the account")

	rootCmd.AddCommand(listCmd)
}
