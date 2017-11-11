package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the version of gowitness",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("gowitness: %s\n", version)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
