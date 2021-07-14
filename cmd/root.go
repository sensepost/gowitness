package cmd

import (
	"fmt"
	"os"

	"github.com/sensepost/gowitness/chrome"
	"github.com/sensepost/gowitness/lib"
	"github.com/sensepost/gowitness/storage"
	"github.com/spf13/cobra"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	options = lib.NewOptions()
	chrm    = chrome.NewChrome()
	db      = storage.NewDb()
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gowitness",
	Short: "A commandline web screenshot and information gathering tool by @leonjza",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {

		// Setup the logger to use
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "02 Jan 2006 15:04:05"})
		if options.Debug {
			log.Logger = log.Logger.Level(zerolog.DebugLevel)
			log.Logger = log.With().Caller().Logger()
			log.Debug().Msg("debug logging enabed")
		} else {
			log.Logger = log.Logger.Level(zerolog.InfoLevel)
		}
		if options.DisableLogging {
			log.Logger = log.Logger.Level(zerolog.Disabled)
		}

		options.Logger = &log.Logger
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// logging
	rootCmd.PersistentFlags().BoolVar(&options.Debug, "debug", false, "enable debug logging")
	rootCmd.PersistentFlags().BoolVar(&options.DisableLogging, "disable-logging", false, "disable all logging")
	// global
	rootCmd.PersistentFlags().BoolVar(&db.Disabled, "disable-db", false, "disable all database operations")
	rootCmd.PersistentFlags().StringVarP(&db.Path, "db-path", "D", "gowitness.sqlite3", "destination for the gowitness database")
	rootCmd.PersistentFlags().IntVarP(&chrm.ResolutionX, "resolution-x", "X", 1440, "screenshot resolution x")
	rootCmd.PersistentFlags().IntVarP(&chrm.ResolutionY, "resolution-y", "Y", 900, "screenshot resolution y")
	rootCmd.PersistentFlags().IntVar(&chrm.Delay, "delay", 0, "delay in seconds between navigation and screenshot")
	rootCmd.PersistentFlags().StringVar(&chrm.UserAgent, "user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.102 Safari/537.36", "user agent string to use")
	rootCmd.PersistentFlags().StringVarP(&options.ScreenshotPath, "screenshot-path", "P", "screenshots", "store path for screenshots (use . for pwd)")
	rootCmd.PersistentFlags().BoolVarP(&chrm.FullPage, "fullpage", "F", false, "take fullpage screenshots")
	rootCmd.PersistentFlags().Int64Var(&chrm.Timeout, "timeout", 10, "preflight check timeout")
	rootCmd.PersistentFlags().StringVarP(&chrm.ChromePath, "chrome-path", "", "", "path to chrome executable to use")
	rootCmd.PersistentFlags().StringVarP(&chrm.Proxy, "proxy", "p", "", "http/socks5 proxy to use. Use format proto://address:port")
}
