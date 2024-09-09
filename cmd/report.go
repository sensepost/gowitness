package cmd

import (
	"github.com/sensepost/gowitness/internal/ascii"
	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Reporting tools",
	Long:  ascii.LogoHelp(`Reporting tools.`),
}

func init() {
	rootCmd.AddCommand(reportCmd)
}
