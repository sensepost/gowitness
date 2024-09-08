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

var cidrCmdOptions = &readers.CidrReaderOptions{}
var cidrCmd = &cobra.Command{
	Use:   "cidr",
	Short: "Scan cidr targets on a network",
	Long: ascii.LogoHelp(`Scan CIDR targets on a network.

This command takes input CIDR ranges, optional extra ports and other
configuration and generates permutations to scan for web services to
screenshot. URL schemes are automatically added as 'http://' and 'https://'
unless either the --no-http or --no-https flags are present.

By default, this command will scan targets sequentially. If the --random
flag is set, targets will go through a shuffling phase before scanning
starts. This is useful in places where scanning too many ports in sequence
may trigger port scanning related alerts.

Note: By default, no metadata is saved except for screenshots that are
stored in the configured --screenshot-path. For later parsing (i.e., using
the gowitness reporting feature), you need to specify where to write results
(db, csv, jsonl) using the --write-* set of flags. See --help for available
flags.`),
	Example: `  Scan multiple CIDR's, using the default ports which are 80 and 443:
   $ gowitness scan cidr --cidr 192.168.0.0/24 --cidr 10.0.50.0/24
  Scan a CIDR (specified using the shorthand -c) with additional ports:
   $ gowitness scan cidr -c 10.0.50.0/24 --port 8888 --port 8443
  Scan multiple CIDR's, using the addition medium set of ports, ignoring http URI's:
   $ gowitness scan cidr -c 172.16.1.0/24 -c 10.10.10.0/24 --no-http --ports-medium
  Scan a CIDR, logging scan errors (can be verbose!) and using 20 'threads':
   $ gowitness scan cidr -t 20 --log-scan-errors -c 10.20.20.0/28`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if cidrCmdOptions.Source == "" && len(cidrCmdOptions.Cidrs) == 0 {
			return errors.New("need targets to scan via either a --cidr-file for --cidr")
		}

		if cidrCmdOptions.Source != "" && !islazy.FileExists(cidrCmdOptions.Source) {
			return errors.New("source is not readable")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("starting cidr scanning", "file", cidrCmdOptions.Source, "cidrs", cidrCmdOptions.Cidrs)

		reader := readers.NewCidrReader(cidrCmdOptions)
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
	scanCmd.AddCommand(cidrCmd)

	cidrCmd.Flags().StringVarP(&cidrCmdOptions.Source, "cidr-file", "f", "", "A file with target CIDR's to scan. Use - for stdin")
	cidrCmd.Flags().BoolVar(&cidrCmdOptions.NoHTTP, "no-http", false, "Do not add 'http://' to targets where missing")
	cidrCmd.Flags().BoolVar(&cidrCmdOptions.NoHTTPS, "no-https", false, "Do not add 'https://' to targets where missing")
	cidrCmd.Flags().StringSliceVarP(&cidrCmdOptions.Cidrs, "cidr", "c", []string{}, "A network CIDR to scan. Supports multiple --cidr flags")
	cidrCmd.Flags().IntSliceVarP(&cidrCmdOptions.Ports, "port", "p", []int{80, 443}, "Ports on targets to scan. Supports multiple --port flags")
	cidrCmd.Flags().BoolVar(&cidrCmdOptions.PortsSmall, "ports-small", false, "Include a small ports list when scanning targets")
	cidrCmd.Flags().BoolVar(&cidrCmdOptions.PortsMedium, "ports-medium", false, "Include a medium ports list when scanning targets")
	cidrCmd.Flags().BoolVar(&cidrCmdOptions.PortsLarge, "ports-large", false, "Include a large ports list when scanning targets")
	cidrCmd.Flags().BoolVar(&cidrCmdOptions.Random, "random", false, "Randomize scan targets")
}
