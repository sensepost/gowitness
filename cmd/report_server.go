package cmd

import (
	"github.com/sensepost/gowitness/internal/ascii"
	"github.com/sensepost/gowitness/web"
	"github.com/spf13/cobra"
)

var serverCmdFlags = struct {
	Host           string
	Port           int
	DbUri          string
	ScreenshotPath string
}{}
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the web user interface",
	Long: ascii.LogoHelp(ascii.Markdown(`
# report server

Start the web user interface.`)),
	Example: ascii.Markdown(`
- gowitness report server
- gowitness report server --port 8080 --db-uri /tmp/gowitness.sqlite3
- gowitness report server --screenshot-path /tmp/screenshots`),
	Run: func(cmd *cobra.Command, args []string) {
		server := web.NewServer(
			serverCmdFlags.Host,
			serverCmdFlags.Port,
			serverCmdFlags.DbUri,
			serverCmdFlags.ScreenshotPath,
		)
		server.Run()
	},
}

func init() {
	reportCmd.AddCommand(serverCmd)

	serverCmd.Flags().StringVar(&serverCmdFlags.Host, "host", "127.0.0.1", "The host address to bind the webserver to")
	serverCmd.Flags().IntVar(&serverCmdFlags.Port, "port", 7171, "The port to start the web server on")
	serverCmd.Flags().StringVar(&serverCmdFlags.DbUri, "db-uri", "sqlite://gowitness.sqlite3", "The database URI to use. Supports SQLite, Postgres, and MySQL (e.g., postgres://user:pass@host:port/db)")
	serverCmd.Flags().StringVar(&serverCmdFlags.ScreenshotPath, "screenshot-path", "./screenshots", "The path where screenshots are stored")
}
