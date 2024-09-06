package cmd

import (
	"fmt"

	"github.com/sensepost/gowitness/internal/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Get gowitness version",
	Run: func(cmd *cobra.Command, args []string) {
		version, gitHash, buildEnv := version.Get()
		fmt.Printf("gowitness: %s\ngit hash: %s\nbuild env: %s\n", version, gitHash, buildEnv)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
