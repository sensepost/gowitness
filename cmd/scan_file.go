package cmd

import (
	"errors"

	"github.com/sensepost/gowitness/internal/islazy"
	"github.com/sensepost/gowitness/pkg/log"
	"github.com/sensepost/gowitness/pkg/readers"
	"github.com/sensepost/gowitness/pkg/runner"
	"github.com/spf13/cobra"
)

var fileCmdOptions = &readers.FileReaderOptions{}
var fileCmd = &cobra.Command{
	Use:   "file",
	Short: "Scan targets sourced from a file",
	Long:  `Scan targets sourced from a file`,
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
