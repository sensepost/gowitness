package cmd

import (
	"errors"

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
	Long:  `Scan targets from a Nessus XML file.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if nessusCmdOptions.Source == "" {
			return errors.New("a source must be specified")
		}

		if !islazy.FileExists(fileCmdOptions.Source) {
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
