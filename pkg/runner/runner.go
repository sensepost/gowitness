package runner

import (
	"context"
	"errors"
	"log/slog"
	"net/url"
	"os"
	"sync"

	wappalyzer "github.com/projectdiscovery/wappalyzergo"
	"github.com/sensepost/gowitness/internal/islazy"
	"github.com/sensepost/gowitness/pkg/models"
	"github.com/sensepost/gowitness/pkg/writers"
)

// Runner is a runner that probes web targets using a driver
type Runner struct {
	Driver     Driver
	Wappalyzer *wappalyzer.Wappalyze

	// options for the Runner to consider
	options Options
	// writers are the result writers to use
	writers []writers.Writer
	// log handler
	log *slog.Logger

	// Targets to scan.
	// This would typically be fed from a gowitness/pkg/reader.
	Targets chan string

	// in case we need to bail
	ctx    context.Context
	cancel context.CancelFunc
}

// New gets a new Runner ready for probing.
// It's up to the caller to call Close() on the runner
func NewRunner(logger *slog.Logger, driver Driver, opts Options, writers []writers.Writer) (*Runner, error) {
	if !opts.Scan.ScreenshotSkipSave {
		screenshotPath, err := islazy.CreateDir(opts.Scan.ScreenshotPath)
		if err != nil {
			return nil, err
		}
		opts.Scan.ScreenshotPath = screenshotPath
		logger.Debug("final screenshot path", "screenshot-path", opts.Scan.ScreenshotPath)
	} else {
		logger.Debug("not saving screenshots to disk")
	}

	// screenshot format check
	if !islazy.SliceHasStr([]string{"jpeg", "png"}, opts.Scan.ScreenshotFormat) {
		return nil, errors.New("invalid screenshot format")
	}

	// javascript file containing javascript to eval on each page.
	// just read it in and set Scan.JavaScript to the value.
	if opts.Scan.JavaScriptFile != "" {
		javascript, err := os.ReadFile(opts.Scan.JavaScriptFile)
		if err != nil {
			return nil, err
		}

		opts.Scan.JavaScript = string(javascript)
	}

	// get a wappalyzer instance
	wap, err := wappalyzer.New()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Runner{
		Driver:     driver,
		Wappalyzer: wap,
		options:    opts,
		writers:    writers,
		Targets:    make(chan string),
		log:        logger,
		ctx:        ctx,
		cancel:     cancel,
	}, nil
}

// runWriters takes a result and passes it to writers
func (run *Runner) runWriters(result *models.Result) error {
	for _, writer := range run.writers {
		if err := writer.Write(result); err != nil {
			return err
		}
	}

	return nil
}

// checkUrl ensures a url is valid
func (run *Runner) checkUrl(target string) error {
	url, err := url.ParseRequestURI(target)
	if err != nil {
		return err
	}

	if !islazy.SliceHasStr(run.options.Scan.UriFilter, url.Scheme) {
		return errors.New("url contains invalid scheme")
	}

	return nil
}

// Run executes the runner, processing targets as they arrive
// in the Targets channel
func (run *Runner) Run() {
	wg := sync.WaitGroup{}

	// will spawn Scan.Theads number of "workers" as goroutines
	for w := 0; w < run.options.Scan.Threads; w++ {
		wg.Add(1)

		// start a worker
		go func() {
			defer wg.Done()
			for {
				select {
				case <-run.ctx.Done():
					return
				case target, ok := <-run.Targets:
					if !ok {
						return
					}

					// validate the target
					if err := run.checkUrl(target); err != nil {
						if run.options.Logging.LogScanErrors {
							run.log.Error("invalid target to scan", "target", target, "err", err)
						}
						continue
					}

					result, err := run.Driver.Witness(target, run)
					if err != nil {
						// is this a chrome not found error?
						var chromeErr *ChromeNotFoundError
						if errors.As(err, &chromeErr) {
							run.log.Error("no valid chrome intallation found", "err", err)
							run.cancel()
							return
						}

						if run.options.Logging.LogScanErrors {
							run.log.Error("failed to witness target", "target", target, "err", err)
						}
						continue
					}

					// assume that status code 0 means there was no information, so
					// don't send anything to writers.
					if result.ResponseCode == 0 {
						if run.options.Logging.LogScanErrors {
							run.log.Error("failed to witness target, status code was 0", "target", target)
						}
						continue
					}

					if err := run.runWriters(result); err != nil {
						run.log.Error("failed to write result for target", "target", target, "err", err)
					}

					run.log.Info("result 🤖", "target", target, "status-code", result.ResponseCode,
						"title", result.Title, "have-screenshot", !result.Failed)

				}
			}

		}()
	}

	wg.Wait()
}

func (run *Runner) Close() {
	// close the driver
	run.Driver.Close()
}
