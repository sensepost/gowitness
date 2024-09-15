package cmd

import (
	"errors"
	"log/slog"

	"github.com/sensepost/gowitness/internal/ascii"
	"github.com/sensepost/gowitness/pkg/log"
	"github.com/sensepost/gowitness/pkg/runner"
	driver "github.com/sensepost/gowitness/pkg/runner/drivers"
	"github.com/sensepost/gowitness/pkg/writers"
	"github.com/spf13/cobra"
)

var scanWriters = []writers.Writer{}
var scanDriver runner.Driver
var scanRunner *runner.Runner

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Perform various scans",
	Long: ascii.LogoHelp(`Perform various scans using sources such as a file,
nmap XML's, Nessus exports or by scanning network CIDR ranges.

By default, gowitness will only take screenshots. However, that is only half
the fun! You can add multiple "writers" that will collect information such as
response codes, content and more. You can specify multiple writers using the
--writer-* flags (see --help).

There are also two "drivers" to choose from, being either "chromedp" or "gorod".
Which to choose depends on a tradeoff. For speed, but a higher chance of failed
screenshots (still recording what gowitness could get using a --writer-*), use
"gorod". If you prefer a *much* better chance of having a screenshot taken, use
"chromedp" (the default). The "chromedp" driver tradeoff is resource usage, however.
`),
	Example: `  Scan targets from a nessus results file, dont prepend http:// to URI targets and filter by port 80:
   $ gowitness scan nessus -f ./scan-results.nessus --port 80
  Scan a targets from a file, skipping http urls and storing network request content as well:
   $ gowitness scan file -f ~/targets.txt --no-http --save-content
  Scan a CIDR, logging scan errors (can be verbose!) and using 20 'threads':
   $ gowitness scan cidr -t 20 --log-scan-errors -c 10.20.20.0/28
  Scan a single target, writing results to a SQLite database and JSON lines:
   $ gowitness scan single -u https://sensepost.com --write-db --write-jsonl`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error

		// annoying quirk, but because im overriding persistentprerun
		// here which overrides the parent it seems.
		// so we need to explicitly call the parents one now.
		if err = rootCmd.PersistentPreRunE(cmd, args); err != nil {
			return err
		}

		// an slog capable logger to use with drivers and runners
		logger := slog.New(log.Logger)

		// configure the driver
		switch opts.Scan.Driver {
		case "gorod":
			scanDriver, err = driver.NewGorod(logger, *opts)
			if err != nil {
				return err
			}
		case "chromedp":
			scanDriver, err = driver.NewChromedp(logger, *opts)
			if err != nil {
				return err
			}
		default:
			return errors.New("invalid scan driver chosen")
		}

		log.Debug("scanning driver started", "driver", opts.Scan.Driver)

		// configure writers that subdommand scanners will pass to
		// a runner instance.
		if opts.Writer.Jsonl {
			w, err := writers.NewJsonWriter(opts.Writer.JsonlFile)
			if err != nil {
				return err
			}
			scanWriters = append(scanWriters, w)
		}

		if opts.Writer.Db {
			w, err := writers.NewDbWriter(opts.Writer.DbURI, opts.Writer.DbDebug)
			if err != nil {
				return err
			}
			scanWriters = append(scanWriters, w)
		}

		if opts.Writer.Csv {
			w, err := writers.NewCsvWriter(opts.Writer.CsvFile)
			if err != nil {
				return err
			}
			scanWriters = append(scanWriters, w)
		}

		if len(scanWriters) == 0 {
			log.Warn("no writers have been configured. only saving screenshots. add writers using --write-* flags")
		}

		// get the runner up. basically all of the subcommands will use this.
		scanRunner, err = runner.NewRunner(logger, scanDriver, *opts, scanWriters)
		if err != nil {
			return err
		}

		return nil
		// TODO: maybe add https://github.com/projectdiscovery/networkpolicy support?
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)

	// logging control for sub commands
	scanCmd.PersistentFlags().BoolVar(&opts.Logging.LogScanErrors, "log-scan-errors", false, "Log scan errors (timeouts, dns errors, etc.) to stderr (warning: can be verbose!)")

	// "threads" && other
	scanCmd.PersistentFlags().StringVarP(&opts.Scan.Driver, "driver", "", "chromedp", "The scan driver to use. Can be one of [gorod, chromedp]")
	scanCmd.PersistentFlags().IntVarP(&opts.Scan.Threads, "threads", "t", 6, "Number of concurrent threads (goroutines) to use")
	scanCmd.PersistentFlags().IntVarP(&opts.Scan.Timeout, "timeout", "T", 60, "Number of seconds before considering a page timed out")
	scanCmd.PersistentFlags().IntVar(&opts.Scan.Delay, "delay", 3, "Number of seconds delay between navigation and screenshotting")
	scanCmd.PersistentFlags().StringSliceVar(&opts.Scan.UriFilter, "uri-filter", []string{"http", "https"}, "Valid URI's to pass to the scanning process")
	scanCmd.PersistentFlags().StringVarP(&opts.Scan.ScreenshotPath, "screenshot-path", "s", "./screenshots", "Path to store screenshots")
	scanCmd.PersistentFlags().StringVar(&opts.Scan.ScreenshotFormat, "screenshot-format", "jpeg", "Format to save screenshots as. Valid formats are: jpeg, png")
	scanCmd.PersistentFlags().BoolVar(&opts.Scan.ScreenshotFullPage, "screenshot-fullpage", false, "Do fullpage screenshots, instead of just the viewport")
	scanCmd.PersistentFlags().StringVar(&opts.Scan.JavaScript, "javascript", "", "A JavaScript function to evaluate on every page, before a screenshot. Note: It must be a JavaScript function! eg: () => console.log('gowitness');")
	scanCmd.PersistentFlags().StringVar(&opts.Scan.JavaScriptFile, "javascript-file", "", "A file containing a JavaScript function to evaluate on every page, before a screenshot. See --javascript")
	scanCmd.PersistentFlags().BoolVar(&opts.Scan.SaveContent, "save-content", false, "Save content from network requests to the configured writers. WARNING: This flag has the potential to make your storage explode in size")

	// chrome options
	scanCmd.PersistentFlags().StringVar(&opts.Chrome.Path, "chrome-path", "", "The path to a Google Chrome binary to use (downloads a platform appropriate binary by default)")
	scanCmd.PersistentFlags().StringVar(&opts.Chrome.Proxy, "chrome-proxy", "", "An http/socks5 proxy server to use. Specify the proxy using this format: proto://address:port")
	scanCmd.PersistentFlags().StringVar(&opts.Chrome.WSS, "chrome-wss-url", "", "A websocket URL to connect to a remote, already running Chrome Devtools instance (i.e., Chrome started with --remote-debugging-port)")
	scanCmd.PersistentFlags().StringVar(&opts.Chrome.UserAgent, "chrome-user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36", "The user-agent string to use")
	scanCmd.PersistentFlags().IntVar(&opts.Chrome.WindowX, "chrome-window-x", 1920, "The Chrome browser window width, in pixels")
	scanCmd.PersistentFlags().IntVar(&opts.Chrome.WindowY, "chrome-window-y", 1080, "The Chrome browser window height, in pixels")
	scanCmd.PersistentFlags().StringSliceVar(&opts.Chrome.Headers, "chrome-header", []string{}, "Extra headers to add to requests. Supports multiple --header flags")

	// write options for scan sub commands
	scanCmd.PersistentFlags().BoolVar(&opts.Writer.Db, "write-db", false, "Write results to a SQLite database")
	scanCmd.PersistentFlags().StringVar(&opts.Writer.DbURI, "write-db-uri", "sqlite://gowitness.sqlite3", "The database URI to use. Supports SQLite, Postgres and MySQL (eg: postgres://user:pass@host:port/db)")
	scanCmd.PersistentFlags().BoolVar(&opts.Writer.DbDebug, "write-db-enable-debug", false, "Enable database query debug logging (warning: verbose!)")
	scanCmd.PersistentFlags().BoolVar(&opts.Writer.Csv, "write-csv", false, "Write results as CSV (has limited columns)")
	scanCmd.PersistentFlags().StringVar(&opts.Writer.CsvFile, "write-csv-file", "gowitness.csv", "The file to write CSV rows to")
	scanCmd.PersistentFlags().BoolVar(&opts.Writer.Jsonl, "write-jsonl", false, "Write results as JSON lines")
	scanCmd.PersistentFlags().StringVar(&opts.Writer.JsonlFile, "write-jsonl-file", "gowitness.jsonl", "The file to write JSON lines to")
}
