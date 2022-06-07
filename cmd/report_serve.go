package cmd

import (
	"github.com/spf13/cobra"
)

// reportServeCmd represents the reportServe command
var reportServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "starts a web server to view screenshot reports (deprecated, use gowitness server instead)",
	Long: `Starts a web server to view screenshot reports.

The global database and screenshot paths should be set to the same as
what they were when a scan was run. The report server also has the ability
to screenshot ad-hoc URLs provided to the submission page.

NOTE: When changing the server address to something other than localhost, make 
sure that only authorised connections can be made to the server port. By default,
access is restricted to localhost to reduce the risk of SSRF attacks against the
host or hosting infrastructure (AWS/Azure/GCP, etc). Consider strict IP filtering
or fronting this server with an authentication aware reverse proxy.

Allowed URLs, by default, need to start with http:// or https://. If you need
this restriction lifted, add the --allow-insecure-uri / -A flag. A word of 
warning though, that also means that someone may request a URL like file:///etc/passwd.
`,
	Run: func(cmd *cobra.Command, args []string) {
		log := options.Logger
		log.Warn().Msg("this command is deprecated. use 'gowitness server' instead")

		serverCmd.Run(cmd, args)
	},
}

func init() {
	reportCmd.AddCommand(reportServeCmd)

	reportServeCmd.Flags().StringVarP(&options.ServerAddr, "address", "a", "localhost:7171", "server listening address")
	reportServeCmd.Flags().BoolVarP(&options.AllowInsecureURIs, "allow-insecure-uri", "A", false, "allow uris that dont start with http(s)")
}
