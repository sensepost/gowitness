package cmd

import (
	"github.com/sensepost/gowitness/internal/ascii"
	"github.com/sensepost/gowitness/pkg/log"
	"github.com/sensepost/gowitness/pkg/writers"
	"github.com/spf13/cobra"
)

var scanCmdWriters = []writers.Writer{}
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Perform various scans",
	Long: ascii.LogoHelp(`Perform various scans using sources such as a file,
nmap XML's, Nessus exports or by scanning network CIDR ranges.`),
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// annoying quirk, but because im overriding persistentprerun
		// here which overrides the parent it seems.
		// so we need to explicitly call the parents one now.
		if err := rootCmd.PersistentPreRunE(cmd, args); err != nil {
			return err
		}

		// TODO: move this somewhere else. it's too elusive where
		// scanWriters come from in subcommands.

		// configure writers that subdommand scanners will pass to
		// a runner instance.
		if opts.Writer.Jsonl {
			w, err := writers.NewJsonWriter(opts.Writer.JsonlFile)
			if err != nil {
				return err
			}
			scanCmdWriters = append(scanCmdWriters, w)
		}

		if opts.Writer.Db {
			w, err := writers.NewDbWriter(opts.Writer.DbURI, opts.Writer.DbDebug)
			if err != nil {
				return err
			}
			scanCmdWriters = append(scanCmdWriters, w)
		}

		if opts.Writer.Csv {
			w, err := writers.NewCsvWriter(opts.Writer.CsvFile)
			if err != nil {
				return err
			}
			scanCmdWriters = append(scanCmdWriters, w)
		}

		if len(scanCmdWriters) == 0 {
			log.Warn("no writers have been configured. only saving screenshots. add writers using --write-* flags")
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
	scanCmd.PersistentFlags().IntVarP(&opts.Scan.Threads, "threads", "t", 6, "Number of concurrent threads (goroutines) to use")
	scanCmd.PersistentFlags().IntVarP(&opts.Scan.Timeout, "timeout", "T", 30, "Number of seconds before considering a page timed out")
	scanCmd.PersistentFlags().StringArrayVar(&opts.Scan.UriFilter, "uri-filter", []string{"http", "https"}, "Valid URI's to pass to the scanning process")
	scanCmd.PersistentFlags().StringVarP(&opts.Scan.ScreenshotPath, "screenshot-path", "s", "./screenshots", "Path to store screenshots")
	scanCmd.PersistentFlags().StringVar(&opts.Scan.UserAgent, "user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36", "The user-agent string to use")

	// write options for scan sub commands
	scanCmd.PersistentFlags().BoolVar(&opts.Writer.Db, "write-db", false, "Write results to a SQLite database")
	scanCmd.PersistentFlags().StringVar(&opts.Writer.DbURI, "write-db-uri", "sqlite://gowitness.sqlite3", "The database URI to use. Supports SQLite and Postgres anhd MySQL (eg: postgres://user:pass@host:port/db)")
	scanCmd.PersistentFlags().BoolVar(&opts.Writer.DbDebug, "write-db-enable-debug", false, "Enable database query debug logging (warning: verbose!)")
	scanCmd.PersistentFlags().BoolVar(&opts.Writer.Csv, "write-csv", false, "Write results as CSV (has limited columns)")
	scanCmd.PersistentFlags().StringVar(&opts.Writer.CsvFile, "write-csv-file", "gowitness.csv", "The file to write CSV rows to")
	scanCmd.PersistentFlags().BoolVar(&opts.Writer.Jsonl, "write-jsonl", false, "Write results as JSON lines")
	scanCmd.PersistentFlags().StringVar(&opts.Writer.JsonlFile, "write-jsonl-file", "gowitness.jsonl", "The file to write JSON lines to")
}
