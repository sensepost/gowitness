package cmd

import (
	"errors"

	"github.com/sensepost/gowitness/internal/ascii"
	"github.com/sensepost/gowitness/internal/islazy"
	"github.com/sensepost/gowitness/pkg/log"
	"github.com/sensepost/gowitness/pkg/readers"
	"github.com/spf13/cobra"
)

var nmapCmdOptions = &readers.NmapReaderOptions{}
var nmapCmd = &cobra.Command{
	Use:   "nmap",
	Short: "Scan targets from an Nmap XML file",
	Long: ascii.LogoHelp(ascii.Markdown(`
# scan nmap

Scan targets from an Nmap XML file.

When performing Nmap scans, specify the -oX nmap.xml flag to store data in an
XML-formatted file that gowitness can parse.

By default, this command will try and screenshot all ports specified in an
nmap.xml results file. That means it will try and do silly things like
screenshot SSH services, which obviously won't work. It's for this reason that
you'd want to specify the ports or services to parse using the --port and
--service / --service-contains flags. For most HTTP-based services, try:
- --service http
- --service http-alt
- --service http-mgmt
- --service http-proxy
- --service https
- --service https-alt

On ports, when specifying --port (can be multiple), target candidates will only
be generated for results that match one of the specified ports. In contrast,
when --exclude-port (can also be multiple) is set, no candidates for that port
will be generated.

**Note**: By default, no metadata is saved except for screenshots that are
stored in the configured --screenshot-path. For later parsing (i.e., using the
gowitness reporting feature), you need to specify where to write results (db,
csv, jsonl) using the _--write-*_ set of flags. See _--help_ for available
flags.`)),
	Example: ascii.Markdown(`
- gowitness scan nmap -f ~/Desktop/targets.xml --write-json --write-db
- gowitness scan nmap -f targets.xml --threads 50 --no-http --service https
- gowitness scan nmap -f /tmp/n.xml --open-only --port 80 --port 443 --port 8080
- gowitness scan nmap -f ~/nmap.xml --open-only --service-contains http`),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if nmapCmdOptions.Source == "" {
			return errors.New("a source must be specified")
		}

		if !islazy.FileExists(nmapCmdOptions.Source) {
			return errors.New("source is not readable")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("starting Nmap file scanning", "file", nmapCmdOptions.Source)

		reader := readers.NewNmapReader(nmapCmdOptions)
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
	scanCmd.AddCommand(nmapCmd)

	nmapCmd.Flags().StringVarP(&nmapCmdOptions.Source, "file", "f", "", "A file with targets to scan. Use - for stdin")
	nmapCmd.Flags().BoolVar(&nmapCmdOptions.NoHTTP, "no-http", false, "Do not add 'http://' to targets where missing")
	nmapCmd.Flags().BoolVar(&nmapCmdOptions.NoHTTPS, "no-https", false, "Do not add 'https://' to targets where missing")
	nmapCmd.Flags().BoolVarP(&nmapCmdOptions.OpenOnly, "open-only", "o", false, "Only scan ports marked as open")
	nmapCmd.Flags().IntSliceVar(&nmapCmdOptions.Ports, "port", []int{}, "A port filter to apply. Supports multiple --port flags")
	nmapCmd.Flags().IntSliceVar(&nmapCmdOptions.ExcludePorts, "exclude-port", []int{}, "A port exclusion filter to apply. Supports multiple --exclude-port flags")
	nmapCmd.Flags().IntSliceVar(&nmapCmdOptions.SkipPorts, "skip-port", []int{}, "Do not scan these ports. Supports multiple --skip-port flags")
	nmapCmd.Flags().StringVar(&nmapCmdOptions.ServiceContains, "service-contains", "", "A service name filter. Will check if service 'contains' this value first")
	nmapCmd.Flags().StringSliceVar(&nmapCmdOptions.Services, "service", []string{}, "A service filter to apply. Supports multiple --service flags")
	nmapCmd.Flags().BoolVar(&nmapCmdOptions.Hostnames, "hostnames", false, "Add hostnames in URL candidates (useful for virtual hosting)")
}
