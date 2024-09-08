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

var nessusCmdOptions = &readers.NessusReaderOptions{}
var nessusCmd = &cobra.Command{
	Use:   "nessus",
	Short: "Scan targets from a Nessus XML file",
	Long: ascii.LogoHelp(`Scan targets from a Nessus XML file.

Targets are parsed out of an exported Nessus scan result in XML format. This
format is typically called "Nessus" format in the export menu.

By default, the parser will search for web services using the following rules:
  - Plugin Name Contains: "Service Detection"
  - Plugin Service Name Contains: "www" or "http"
  - Plugin Output Value Contains: "web server"

With these defaults, the parser should detect most web services from a Nessus
scan export. You can adjust the filters to include more Plugin Names, Service
Names, or Plugin Output filters using the --service-name, --plugin-output and
--plugin-name flags.

Including the --hostnames flag will have the parser add a scan target based on
any hostname information found in a matched result.

Note: By default, no metadata is saved except for screenshots that are
stored in the configured --screenshot-path. For later parsing (i.e., using
the gowitness reporting feature), you need to specify where to write results
(db, csv, jsonl) using the --write-* set of flags. See --help for available
flags.`),
	Example: `  Scan targets from a file:
   $ gowitness scan nessus -f ~/Desktop/scan-results.nessus
  Scan targets from a file, using 50 'threads' and filtering plugin output by the word 'server':
   $ gowitness scan nessus -f scan-results.nessus --threads 50 --plugin-output server
  Scan targets from a file, dont prepend http:// to URI targets and filter by port 80:
   $ gowitness scan nessus -f ./scan-results.nessus --port 80`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if nessusCmdOptions.Source == "" {
			return errors.New("a source must be specified")
		}

		if !islazy.FileExists(nessusCmdOptions.Source) {
			return errors.New("source is not readable")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("starting nessus file scanning", "file", nessusCmdOptions.Source)

		reader := readers.NewNessusReader(nessusCmdOptions)

		runner, err := runner.New(*opts, scanCmdWriters)
		if err != nil {
			log.Error("could not get a runner", "err", err)
			return
		}
		defer runner.Close()

		go func() {
			reader.Read(runner.Targets)
		}()

		runner.Run()
	},
}

func init() {
	scanCmd.AddCommand(nessusCmd)

	nessusCmd.Flags().StringVarP(&nessusCmdOptions.Source, "file", "f", "", "A file with targets to scan. Use - for stdin")
	nessusCmd.Flags().BoolVar(&nessusCmdOptions.NoHTTP, "no-http", false, "Do not add 'http://' to targets where missing")
	nessusCmd.Flags().BoolVar(&nessusCmdOptions.NoHTTPS, "no-https", false, "Do not add 'https://' to targets where missing")
	nessusCmd.Flags().BoolVar(&nessusCmdOptions.Hostnames, "hostnames", false, "Enable hostname scanning")
	nessusCmd.Flags().StringSliceVar(&nessusCmdOptions.Services, "service-name", []string{"www", "http"}, "Service name is filter. Supports multiple --service flags")
	nessusCmd.Flags().StringSliceVar(&nessusCmdOptions.PluginOutputs, "plugin-output", []string{"web server"}, "Plugin output contains filter. Supports multiple --plugin-output flags")
	nessusCmd.Flags().StringSliceVar(&nessusCmdOptions.PluginNames, "plugin-name", []string{"Service Detection"}, "Plugin name is filter. Supports multiple --plugin-name flags")
	nessusCmd.Flags().IntSliceVar(&nessusCmdOptions.Ports, "port", []int{}, "Port filter. Supports multiple --port flags")
}
