package runner

import (
	"errors"
	"net/url"
	"os"
	"sync"

	wappalyzer "github.com/projectdiscovery/wappalyzergo"
	"github.com/sensepost/gowitness/internal/islazy"
	"github.com/sensepost/gowitness/pkg/log"
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

	// Targets to scan.
	// This would typically be fed from a gowitness/pkg/reader.
	Targets chan string
}

// New gets a new Runner ready for probing.
// It's up to the caller to call Close() on the runner
func NewRunner(driver Driver, opts Options, writers []writers.Writer) (*Runner, error) {
	screenshotPath, err := islazy.CreateDir(opts.Scan.ScreenshotPath)
	if err != nil {
		return nil, err
	}
	opts.Scan.ScreenshotPath = screenshotPath
	log.Debug("final screenshot path", "screenshot-path", opts.Scan.ScreenshotPath)

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

	return &Runner{
		Driver:     driver,
		Wappalyzer: wap,
		options:    opts,
		writers:    writers,
		Targets:    make(chan string),
	}, nil
}

// InvokeWriters takes a result and passes it to writers
func (run *Runner) InvokeWriters(result *models.Result) error {
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
			for target := range run.Targets {
				log.Info("working with target")
				// validate the target
				if err := run.checkUrl(target); err != nil {
					if run.options.Logging.LogScanErrors {
						log.Error("invalid target to scan", "target", target, "err", err)
					}
					continue
				}

				// process the target
				// TODO: bubble an error up from witness()
				run.Driver.Witness(target, run)
			}
		}()
	}

	wg.Wait()
}

func (run *Runner) Close() {
	// close the driver
	run.Driver.Close()
}
