package cmd

import (
	"github.com/sensepost/gowitness/internal/ascii"
	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Work with gowitness reports",
	Long: ascii.LogoHelp(ascii.Markdown(`
# report

Work with gowitness reports.
`)),
}

func init() {
	rootCmd.AddCommand(reportCmd)
}
