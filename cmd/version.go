package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	gitHash string
	goVer   string
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the version of gowitness",
	Run: func(cmd *cobra.Command, args []string) {

		if gitHash == "" {
			gitHash = "dev"
		}

		if goVer == "" {
			goVer = "dev"
		}

		fmt.Printf("gowitness: %s\n", version)
		fmt.Printf("\ngit hash: %s\ngo version: %s\n", gitHash, goVer)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
