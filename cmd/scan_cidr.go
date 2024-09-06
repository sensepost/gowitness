package cmd

import (
	"errors"

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
	Long:  `Scan cidr targets on a network`,
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
