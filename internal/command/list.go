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

	// Command
	listCmd = &cobra.Command{
		Use:   "list [org]",
		Short: "List information about certain resources",
		Long: `The list command can be used to view the list of certain resources
such as accounts and organizations.`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			orgName := args[0]

			log.Debug("Loading org", "name", orgName)

			org, err := NewAwsOrgFromConfig(orgName)

			if err != nil {
				log.Fatal(err)
			}

			pattern, err := regexp.Compile(filter)

			if err != nil {
				log.Fatal(err)
			}

			writer := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
			defer writer.Flush()

			for _, account := range org.Accounts {
				if !pattern.MatchString(account.Name) {
					continue
				}

				line := account.Name

				if printId {
					line = account.ID + "\t" + account.Name
				}

				if printEmail {
					line += "\t" + account.Email
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

	rootCmd.AddCommand(listCmd)
}
