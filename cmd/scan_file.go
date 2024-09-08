package cmd

import (
	"errors"

	"github.com/sensepost/gowitness/internal/ascii"
	"github.com/sensepost/gowitness/internal/islazy"
	"github.com/sensepost/gowitness/pkg/log"
	"github.com/sensepost/gowitness/pkg/readers"
	"github.com/sensepost/gowitness/pkg/runner"
	"github.com/spf13/cobra"
)

var fileCmdOptions = &readers.FileReaderOptions{}
var fileCmd = &cobra.Command{
	Use:   "file",
	Short: "Scan targets sourced from a file or stdin",
	Long: ascii.LogoHelp(`Scan targets sourced from a file or stdin.

This command will check the structure of a target URL to ensure that a
protocol is defined. If it is not set, it will prepend 'http://' and
'https://'. You can disable either using the --no-http / --no-https flags.

URLs in the source file should be newline separated. Invalid URLs are
simply ignored.

Note: By default, no metadata is saved except for screenshots that are
stored in the configured --screenshot-path. For later parsing (i.e., using
the gowitness reporting feature), you need to specify where to write results
(db, csv, jsonl) using the --write-* set of flags. See --help for available
flags.`),
	Example: `  Scan targets from a file:
   $ gowitness scan file -f ~/Desktop/targets.txt
  Scan targets from a file, using 50 'threads':
   $ gowitness scan file -f targets.txt --threads 50
  Scan targets from a file piped in via stdin:
   $ cat urls.txt | gowitness scan file -f -
  Scan targets from a file that is first shuffled using the 'shuf' command.
  This can also wont prepend http:// to any urls that need don't have a URI:
   $ gowitness scan file -f <( shuf domains.txt ) --no-http`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if fileCmdOptions.Source == "" {
			return errors.New("a source must be specified")
		}

		if !islazy.FileExists(fileCmdOptions.Source) {
			return errors.New("source is not readable")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("starting file scanning", "file", fileCmdOptions.Source)

		reader := readers.NewFileReader(fileCmdOptions)
		runner, err := runner.New(*opts, scanCmdWriters)
		if err != nil {
			log.Error("could not get a runner", "err", err)
			return
		}
		defer runner.Close()

		go func() {
			if err := reader.Read(runner.Targets); err != nil {
				log.Error("error in reader.Read", "err", err)
				return
			}
		}()

		runner.Run()
	},
}

func init() {
	scanCmd.AddCommand(fileCmd)

	fileCmd.Flags().StringVarP(&fileCmdOptions.Source, "file", "f", "", "A file with targets to scan. Use - for stdin")
	fileCmd.Flags().BoolVar(&fileCmdOptions.NoHTTP, "no-http", false, "Do not add 'http://' to targets where missing")
	fileCmd.Flags().BoolVar(&fileCmdOptions.NoHTTPS, "no-https", false, "Do not add 'https://' to targets where missing")
}
