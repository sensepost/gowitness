package cmd

import (
	"github.com/sensepost/gowitness/pkg/log"
	"github.com/sensepost/gowitness/pkg/readers"
	"github.com/sensepost/gowitness/pkg/runner"
	"github.com/spf13/cobra"
)

// nessusCmd represents the nessus command
var nessusCmd = &cobra.Command{
	Use:   "nessus",
	Short: "Scan targets from a Nessus XML file",
	Long:  `Scan targets from a Nessus XML file.`,
	Run: func(cmd *cobra.Command, args []string) {
		source, _ := cmd.Flags().GetString("file")
		nohttp, _ := cmd.Flags().GetBool("no-http")
		nohttps, _ := cmd.Flags().GetBool("no-https")
		hostnames, _ := cmd.Flags().GetBool("hostnames")
		services, _ := cmd.Flags().GetStringSlice("service-name")
		pluginOutputs, _ := cmd.Flags().GetStringSlice("plugin-output")
		pluginNames, _ := cmd.Flags().GetStringSlice("plugin-name")
		ports, _ := cmd.Flags().GetIntSlice("port")
		log.Debug("starting nessus file scanning", "file", source)

		reader := readers.NewNessusReader(source, &readers.NessusReaderOptions{
			NoHTTP:        nohttp,
			NoHTTPS:       nohttps,
			Hostnames:     hostnames,
			Services:      services,
			PluginOutputs: pluginOutputs,
			PluginNames:   pluginNames,
			Ports:         ports,
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
	scanCmd.AddCommand(nessusCmd)

	nessusCmd.Flags().StringP("file", "f", "", "An Nmap XML file with targets to scan")

	// options
	nessusCmd.Flags().Bool("no-https", false, "Do not add 'https://' to targets where missing")
	nessusCmd.Flags().Bool("no-http", false, "Do not add 'http://' to targets where missing")
	nessusCmd.Flags().Bool("hostnames", false, "Enable hostname scanning")
	nessusCmd.Flags().StringSlice("service-name", []string{"www", "http"}, "Service name is filter. Supports multiple --service flags")
	nessusCmd.Flags().StringSlice("plugin-output", []string{"web server"}, "Plugin output contains filter. Supports multiple --plugin-output flags")
	nessusCmd.Flags().StringSlice("plugin-name", []string{"Service Detection"}, "Plugin name is filter. Supports multiple --plugin-name flags")
	nessusCmd.Flags().IntSlice("port", []int{}, "Port filter. Supports multiple --port flags")
}
