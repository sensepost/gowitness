package cmd

import (
	"errors"

	"github.com/sensepost/gowitness/internal/ascii"
	"github.com/spf13/cobra"
)

var singleCmdOptions = struct {
	URL string
}{}

var singleCmd = &cobra.Command{
	Use:   "single",
	Short: "Scan a single URL target",
	Long: ascii.LogoHelp(`Scan a single URL target.

Note: By default, no metadata is saved except for screenshots that are
stored in the configured --screenshot-path. For later parsing (i.e., using
the gowitness reporting feature), you need to specify where to write results
(db, csv, jsonl) using the --write-* set of flags. See --help for available
flags.`),
	Example: `  Scan a single target:
   $ gowitness scan single -u https://sensepost.com`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if singleCmdOptions.URL == "" {
			return errors.New("a url must be specified")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		url, _ := cmd.Flags().GetString("url")

		go func() {
			scanRunner.Targets <- url
			close(scanRunner.Targets)
		}()

		scanRunner.Run()
		scanRunner.Close()
	},
}

func init() {
	scanCmd.AddCommand(singleCmd)

	singleCmd.Flags().StringVarP(&singleCmdOptions.URL, "url", "u", "", "The target to screenshot")
}
