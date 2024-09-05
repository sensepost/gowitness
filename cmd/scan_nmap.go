package cmd

import (
	"github.com/sensepost/gowitness/internal/validators"
	"github.com/sensepost/gowitness/pkg/log"
	"github.com/sensepost/gowitness/pkg/readers"
	"github.com/sensepost/gowitness/pkg/runner"
	"github.com/spf13/cobra"
)

// nmapCmd represents the nmap command
var nmapCmd = &cobra.Command{
	Use:   "nmap",
	Short: "Scan targets from an Nmap XML file",
	Long:  `Scan targets from an Nmap XML file`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := validators.ValidateScanNmapCmd(cmd); err != nil {
			return err
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		source, _ := cmd.Flags().GetString("file")
		nohttp, _ := cmd.Flags().GetBool("no-http")
		nohttps, _ := cmd.Flags().GetBool("no-https")
		openOnly, _ := cmd.Flags().GetBool("open-only")
		ports, _ := cmd.Flags().GetIntSlice("port")
		skipPorts, _ := cmd.Flags().GetIntSlice("skip-port")
		serviceContains, _ := cmd.Flags().GetString("service")
		service, _ := cmd.Flags().GetStringSlice("service")
		hostnames, _ := cmd.Flags().GetBool("hostname")
		log.Debug("starting nmap file scanning", "file", source)

		reader := readers.NewNmapReader(source, &readers.NmapReaderOptions{
			NoHTTP:          nohttp,
			NoHTTPS:         nohttps,
			OpenOnly:        openOnly,
			Ports:           ports,
			SkipPorts:       skipPorts,
			ServiceContains: serviceContains,
			Service:         service,
			Hostnames:       hostnames,
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
	scanCmd.AddCommand(nmapCmd)

	nmapCmd.Flags().StringP("file", "f", "", "An Nmap XML file with targets to scan")

	// options
	nmapCmd.Flags().Bool("no-https", false, "Do not add 'https://' to targets where missing")
	nmapCmd.Flags().Bool("no-http", false, "Do not add 'http://' to targets where missing")
	nmapCmd.Flags().BoolP("open-only", "o", false, "Only scan ports marked as open")
	nmapCmd.Flags().IntSlice("port", []int{}, "A port filter to apply. Suports multiple --port flags")
	nmapCmd.Flags().IntSlice("skip-port", []int{}, "Do not scan these ports. Suports multiple --skip-port flags")
	nmapCmd.Flags().String("service-contains", "", "A service name filter. Will check if service 'contains' this value first")
	nmapCmd.Flags().StringSlice("service", []string{}, "A service filter to apply. Supports multiple --service flags")
	nmapCmd.Flags().Bool("hostnames", false, "Add hostnames in URL candidates (useful for virtual hosting)")
}
