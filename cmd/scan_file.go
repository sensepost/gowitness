package cmd

import (
	"github.com/sensepost/gowitness/internal/validators"
	"github.com/sensepost/gowitness/pkg/log"
	"github.com/sensepost/gowitness/pkg/readers"
	"github.com/sensepost/gowitness/pkg/runner"
	"github.com/spf13/cobra"
)

// fileCmd represents the file command
var fileCmd = &cobra.Command{
	Use:   "file",
	Short: "Scan targets sourced from a file",
	Long:  `Scan targets sourced from a file`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := validators.ValidateScanFileCmd(cmd); err != nil {
			return err
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		source, _ := cmd.Flags().GetString("file")
		nohttp, _ := cmd.Flags().GetBool("no-http")
		nohttps, _ := cmd.Flags().GetBool("no-https")
		log.Debug("starting file scanning", "file", source)

		reader := readers.NewFileReader(source, &readers.FileReaderOptions{
			NoHTTP:  nohttp,
			NoHTTPS: nohttps,
		})

		runner, err := runner.New(*opts, scanWriters)
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
	scanCmd.AddCommand(fileCmd)

	fileCmd.Flags().StringP("file", "f", "", "A file with targets to scan. Use - for stdin")
	fileCmd.Flags().BoolP("no-https", "", false, "Do not add 'https://' to targets where missing")
	fileCmd.Flags().BoolP("no-http", "", false, "Do not add 'http://' to targets where missing")
}
