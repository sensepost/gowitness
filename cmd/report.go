package cmd

import (
	"github.com/spf13/cobra"
)

// reportCmd represents the report command
var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Work with gowitness reports",
	Long:  `Work with gowitness reports`,
}

func init() {
	RootCmd.AddCommand(reportCmd)
}
