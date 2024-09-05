package cmd

import (
	"errors"

	"github.com/sensepost/gowitness/internal/islazy"
	"github.com/sensepost/gowitness/pkg/log"
	"github.com/sensepost/gowitness/pkg/readers"
	"github.com/sensepost/gowitness/pkg/runner"
	"github.com/spf13/cobra"
)

var nmapCmdOptions = &readers.NmapReaderOptions{}
var nmapCmd = &cobra.Command{
	Use:   "nmap",
	Short: "Scan targets from an Nmap XML file",
	Long:  `Scan targets from an Nmap XML file`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if nmapCmdOptions.Source == "" {
			return errors.New("a source must be specified")
		}

		if !islazy.FileExists(fileCmdOptions.Source) {
			return errors.New("source is not readable")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("starting nmap file scanning", "file", nmapCmdOptions.Source)

		reader := readers.NewNmapReader(nmapCmdOptions)
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
	scanCmd.AddCommand(nmapCmd)

	nmapCmd.Flags().StringVarP(&nmapCmdOptions.Source, "file", "f", "", "A file with targets to scan. Use - for stdin")
	nmapCmd.Flags().BoolVar(&nmapCmdOptions.NoHTTP, "no-http", false, "Do not add 'http://' to targets where missing")
	nmapCmd.Flags().BoolVar(&nmapCmdOptions.NoHTTPS, "no-https", false, "Do not add 'https://' to targets where missing")
	nmapCmd.Flags().BoolVarP(&nmapCmdOptions.OpenOnly, "open-only", "o", false, "Only scan ports marked as open")
	nmapCmd.Flags().IntSliceVar(&nmapCmdOptions.Ports, "port", []int{}, "A port filter to apply. Suports multiple --port flags")
	nmapCmd.Flags().IntSliceVar(&nmapCmdOptions.SkipPorts, "skip-port", []int{}, "Do not scan these ports. Suports multiple --skip-port flags")
	nmapCmd.Flags().StringVar(&nmapCmdOptions.ServiceContains, "service-contains", "", "A service name filter. Will check if service 'contains' this value first")
	nmapCmd.Flags().StringSliceVar(&nmapCmdOptions.Services, "service", []string{}, "A service filter to apply. Supports multiple --service flags")
	nmapCmd.Flags().BoolVar(&nmapCmdOptions.Hostnames, "hostnames", false, "Add hostnames in URL candidates (useful for virtual hosting)")
}
