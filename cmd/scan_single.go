package cmd

import (
	"errors"

	"github.com/sensepost/gowitness/internal/ascii"
	"github.com/sensepost/gowitness/pkg/log"
	"github.com/sensepost/gowitness/pkg/runner"
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
		runner, err := runner.New(*opts, scanCmdWriters)
		if err != nil {
			log.Error("could not get a runner", "err", err)
			return
		}
		defer runner.Close()

		go func() {
			runner.Targets <- url
			close(runner.Targets)
		}()

		runner.Run()
	},
}

func init() {
	scanCmd.AddCommand(singleCmd)

	singleCmd.Flags().StringVarP(&singleCmdOptions.URL, "url", "u", "", "The target to screenshot")
}
