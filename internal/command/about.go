package command

import (
	"fmt"

	"github.com/schubergphilis/grawsp/internal/meta"
	"github.com/spf13/cobra"
)

var (
	// Command
	aboutCmd = &cobra.Command{
		Use:   "about",
		Short: "Information about the development of grawsp",
		Long: `This command outputs information related to the development of
grawsp, such as build, release and developers.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Version: ", meta.Version)
			fmt.Println("Build Time: ", meta.BuildTime)
			fmt.Println("Commit: ", meta.Commit)
		},
	}
)

func init() {
	rootCmd.AddCommand(aboutCmd)
}
