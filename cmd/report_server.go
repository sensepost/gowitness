package cmd

import (
	"github.com/sensepost/gowitness/internal/ascii"
	"github.com/sensepost/gowitness/web"
	"github.com/spf13/cobra"
)

// serverCmd represents the server command
var serverCmdFlags = struct {
	Port           int
	DbUri          string
	ScreenshotPath string
}{}
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the report web server user interface",
	Long:  ascii.LogoHelp(`Start the report web server user interface.`),
	Run: func(cmd *cobra.Command, args []string) {
		server := web.NewServer(
			serverCmdFlags.Port,
			serverCmdFlags.DbUri,
			serverCmdFlags.ScreenshotPath,
		)
		server.Run()
	},
}

func init() {
	reportCmd.AddCommand(serverCmd)

	serverCmd.Flags().IntVar(&serverCmdFlags.Port, "port", 7171, "The port to start the web server on")
	serverCmd.PersistentFlags().StringVar(&serverCmdFlags.DbUri, "db-uri", "sqlite://gowitness.sqlite3", "The database URI to use. Supports SQLite and Postgres and MySQL (eg: postgres://user:pass@host:port/db)")
	serverCmd.PersistentFlags().StringVar(&serverCmdFlags.ScreenshotPath, "screenshot-path", "./screenshots", "The path where screenshots are stored")
}
