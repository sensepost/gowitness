package cmd

import (
	"errors"

	"github.com/sensepost/gowitness/internal/ascii"
	"github.com/sensepost/gowitness/internal/islazy"
	"github.com/sensepost/gowitness/pkg/log"
	"github.com/sensepost/gowitness/pkg/readers"
	"github.com/spf13/cobra"
)

var fileCmdOptions = &readers.FileReaderOptions{}
var fileCmd = &cobra.Command{
	Use:   "file",
	Short: "Scan targets sourced from a file or stdin",
	Long: ascii.LogoHelp(ascii.Markdown(`
# scan file

Scan targets sourced from a file or stdin.

## description

This command will check the structure of a target URL to ensure that a protocol
is defined. If it is not set, it will prepend 'http://' and 'https://'. You can
disable either using the --no-http / --no-https flags.

URLs in the source file should be newline-separated. Invalid URLs are simply
ignored.

If any ports are added (via --port or one of the ports collections), then URL
candidates will also be generated with the port section specified.

**Note**: By default, no metadata is saved except for screenshots that are
stored in the configured --screenshot-path. For later parsing (i.e., using the
gowitness reporting feature), you need to specify where to write results (db,
csv, jsonl) using the _--write-*_ set of flags. See _--help_ for available
flags.`)),
	Example: ascii.Markdown(`
- gowitness scan file -f ~/Desktop/targets.txt --write-jsonl
- gowitness scan file -f targets.txt --threads 50 --write-db
- cat urls.txt | gowitness scan file -f - --write-csv
- gowitness scan file -f <( shuf domains.txt ) --no-http
`),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if fileCmdOptions.Source == "" {
			return errors.New("a source must be specified")
		}

		if fileCmdOptions.Source != "-" && !islazy.FileExists(fileCmdOptions.Source) {
			return errors.New("source is not readable")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("starting file scanning", "file", fileCmdOptions.Source)

		reader := readers.NewFileReader(fileCmdOptions)
		go func() {
			if err := reader.Read(scanRunner.Targets); err != nil {
				log.Error("error in reader.Read", "err", err)
				return
			}
		}()

		scanRunner.Run()
		scanRunner.Close()
	},
}

func init() {
	scanCmd.AddCommand(fileCmd)

	fileCmd.Flags().StringVarP(&fileCmdOptions.Source, "file", "f", "", "A file with targets to scan. Use - for stdin")
	fileCmd.Flags().BoolVar(&fileCmdOptions.NoHTTP, "no-http", false, "Do not add 'http://' to targets where missing")
	fileCmd.Flags().BoolVar(&fileCmdOptions.NoHTTPS, "no-https", false, "Do not add 'https://' to targets where missing")
	fileCmd.Flags().IntSliceVarP(&fileCmdOptions.Ports, "port", "p", []int{80, 443}, "Ports on targets to scan. Supports multiple --port flags")
	fileCmd.Flags().BoolVar(&fileCmdOptions.PortsSmall, "ports-small", false, "Include a small ports list when scanning targets")
	fileCmd.Flags().BoolVar(&fileCmdOptions.PortsMedium, "ports-medium", false, "Include a medium ports list when scanning targets")
	fileCmd.Flags().BoolVar(&fileCmdOptions.PortsLarge, "ports-large", false, "Include a large ports list when scanning targets")
}
