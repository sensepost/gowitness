package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	homedir "github.com/mitchellh/go-homedir"
	chrm "github.com/sensepost/gowitness/chrome"
	log "github.com/sirupsen/logrus"

	"github.com/sensepost/gowitness/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile    string
	chrome     chrm.Chrome
	db         storage.Storage
	dbLocation string

	// logging
	logLevel  string
	logFormat string

	// 'global' flags
	waitTimeout      int
	resolution       string
	chromeTimeout    int
	chromeTimeBudget int
	chromePath       string
	userAgent        string
	disableDb        bool
	chromeArgvs      []string

	// screenshot command flags
	screenshotURL         string
	screenshotDestination string

	// file scanner command flags
	sourceFile  string
	maxThreads  int
	prefixHTTP  bool
	prefixHTTPS bool

	// nmap scanner command flags
	nmapFile      string
	scanHostnames bool
	nmapServices  []string
	nmapPorts     []int

	// range scanner command flags
	scanCidr           []string
	scanFileCidr       string
	scanPorts          string
	skipHTTP           bool // used in nmap command too
	skipHTTPS          bool // used in nmap command too
	randomPermutations bool
	appendURI          string
	appendURIFile      string

	// report generate command
	reportFileName    string
	reportChunks      int
	perceptionSort    bool
	statusCodeSort    bool
	titleSort         bool
	filterStatusCodes []int
	ignoreFailed      bool

	// report list command
	colWidth int

	// execution time
	startTime = time.Now()

	// version
	version = "1.3.0"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "gowitness",
	Short: "A commandline web screenshot and information gathering tool by @leonjza",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		initLogging()
		validateFlags()

		// Init Google Chrome
		chrome = chrm.Chrome{
			Resolution:       resolution,
			ChromeTimeout:    chromeTimeout,
			ChromeTimeBudget: chromeTimeBudget,
			Path:             chromePath,
			UserAgent:        userAgent,
			Argvs:            chromeArgvs,
		}
		chrome.Setup()

		// Setup the destination directory
		if err := chrome.SetScreenshotPath(screenshotDestination); err != nil {
			log.WithField("error", err).Fatal("Error in setting destination screenshot path.")
		}

		db = storage.Storage{}

		if disableDb {
			db.Enabled = false
		} else {
			db.Enabled = true
			db.Open(dbLocation)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		log.WithField("error", err).Fatal("exited with error")
		os.Exit(1)
	}
}

func init() {
	// cobra.OnInitialize(initConfig)

	// logging
	RootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "one of debug, info, warn, error, or fatal")
	RootCmd.PersistentFlags().StringVar(&logFormat, "log-format", "text", "specify output (text or json)")

	// Global flags
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gowitness.yaml)")
	RootCmd.PersistentFlags().IntVarP(&waitTimeout, "timeout", "T", 3, "Time in seconds to wait for a HTTP connection")
	RootCmd.PersistentFlags().IntVarP(&chromeTimeout, "chrome-timeout", "", 90, "Time in seconds to wait for Google Chrome to finish a screenshot")
	RootCmd.PersistentFlags().IntVarP(&chromeTimeBudget, "chrome-time-budget", "", 2, "Time in seconds to wait for pending network requests when loading a page in Google Chrome")
	RootCmd.PersistentFlags().StringVarP(&chromePath, "chrome-path", "", "", "Full path to the Chrome executable to use. By default, gowitness will search for Google Chrome")
	RootCmd.PersistentFlags().StringVarP(&userAgent, "user-agent", "", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/77.0.3865.75 Safari/537.36", "Alternate UserAgent string to use for Google Chrome")
	RootCmd.PersistentFlags().StringVarP(&resolution, "resolution", "R", "1440,900", "screenshot resolution")
	RootCmd.PersistentFlags().StringVarP(&screenshotDestination, "destination", "d", ".", "Destination directory for screenshots")
	RootCmd.PersistentFlags().BoolVarP(&disableDb, "disable-db", "", false, "Disable database features (wont write a gowitness.db)")
	RootCmd.PersistentFlags().StringVarP(&dbLocation, "db", "D", "gowitness.db", "Destination for the gowitness database")
	RootCmd.PersistentFlags().StringSliceVarP(&chromeArgvs, "chrome-arg", "g", []string{}, "Extra arguments to pass to chrome headless")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)

	} else {

		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".gowitness" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".gowitness")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

// initLogging prepares the logrus logger and format.
// the --log-level and --log-format commandline args lets you
// control what and how logrus outputs data.
func initLogging() {

	switch logLevel {

	case "debug":
		log.SetLevel(log.DebugLevel)

	case "info":
		log.SetLevel(log.InfoLevel)

	case "warn":
		log.SetLevel(log.WarnLevel)

	case "error":
		log.SetLevel(log.ErrorLevel)

	case "fatal":
		log.SetLevel(log.FatalLevel)

	default:
		log.WithField("log-level", logLevel).Warning("invalid log level. defaulting to info.")
		log.SetLevel(log.InfoLevel)
	}

	// Include timestamps in the text format output
	textformat := &log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	}

	switch logFormat {

	case "text":
		log.SetFormatter(textformat)

	case "json":
		log.SetFormatter(new(log.JSONFormatter))

	default:
		log.WithField("log-format", logFormat).Warning("invalid log format. defaulting to text.")
		log.SetFormatter(textformat)
	}
}

// Checks if some of the globally provided arguments are valid.
func validateFlags() {

	// Check screenresolution argument values
	parsedResolution := strings.Split(resolution, ",")
	if len(parsedResolution) != 2 {

		log.WithField("resolution", resolution).Fatal("Invalid resolution value provided")
	}

	if _, err := strconv.Atoi(parsedResolution[0]); err != nil {
		log.WithField("resolution", resolution).Fatal("Failed to parse resolution x value")
	}

	if _, err := strconv.Atoi(parsedResolution[1]); err != nil {
		log.WithField("resolution", resolution).Fatal("Failed to parse resolution y value")
	}

}
