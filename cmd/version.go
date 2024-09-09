package cmd

import (
	"fmt"

	"github.com/sensepost/gowitness/internal/ascii"
	"github.com/sensepost/gowitness/internal/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Get the gowitness version",
	Long:  ascii.LogoHelp(`Get the gowitness version.`),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(ascii.Logo())
		fmt.Printf("\ngowitness: %s\ngit hash: %s\nbuild env: %s\nbuild time: %s\n",
			version.Version, version.GitHash, version.GoBuildEnv, version.GoBuildTime)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
