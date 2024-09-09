package runner

import (
	"bytes"
	"errors"
	"image"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/corona10/goimagehash"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	wappalyzer "github.com/projectdiscovery/wappalyzergo"
	"github.com/sensepost/gowitness/internal/islazy"
	"github.com/sensepost/gowitness/pkg/log"
	"github.com/sensepost/gowitness/pkg/models"
	"github.com/sensepost/gowitness/pkg/writers"
	"github.com/ysmood/gson"
)

// Runner is a runner that probes web targets
type Runner struct {
	// browser is a go-rod browser instance
	browser    *rod.Browser
	wappalyzer *wappalyzer.Wappalyze

	// options for the Runner to consider
	options Options
	// writers are the result writers to use
	writers []writers.Writer

	// Targets to scan.
	// This would typically be fed from a gowitness/pkg/reader.
	Targets chan string
}

// New gets a new Browser ready for probing.
// It's up to the caller to call Close() on the instance.
func New(opts Options, writers []writers.Writer) (*Runner, error) {
	screenshotPath, err := islazy.CreateDir(opts.Scan.ScreenshotPath)
	if err != nil {
		return nil, err
	}
	opts.Scan.ScreenshotPath = screenshotPath
	log.Debug("final screenshot path", "screenshot-path", opts.Scan.ScreenshotPath)

	// get a wappalyzer instance
	wap, err := wappalyzer.New()
	if err != nil {
		return nil, err
	}

	// silly window size check
	if !strings.Contains(opts.Chrome.WindowSize, ",") {
		return nil, errors.New("window size appears to be malformed")
	}

	// screenshot format check
	if !islazy.SliceHasStr([]string{"jpeg", "png"}, opts.Scan.ScreenshotFormat) {
		return nil, errors.New("invalid screenshot format")
	}

	// TODO: configure logging
	// TODO: is root, disable sandbox

	// get chrome ready
	chrmLauncher := launcher.New().
		// https://github.com/GoogleChrome/chrome-launcher/blob/main/docs/chrome-flags-for-tools.md
		Set("disable-features", "MediaRouter").
		Set("disable-client-side-phishing-detection").
		Set("disable-default-apps").
		Set("hide-scrollbars").
		Set("mute-audio").
		Set("no-default-browser-check").
		Set("no-first-run").
		Set("deny-permission-prompts").
		Set("window-size", opts.Chrome.WindowSize)

	// user specified Chrome
	if opts.Chrome.Path != "" {
		chrmLauncher.Bin(opts.Chrome.Path)
	}

	// proxy
	if opts.Chrome.Proxy != "" {
		chrmLauncher.Proxy(opts.Chrome.Proxy)
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

	url, err := chrmLauncher.Launch()
	if err != nil {
		return nil, err
	}

	browser := rod.New().ControlURL(url).MustConnect().MustIgnoreCertErrors(true)
	log.Debug("got a browser up", "control-url", url)

	return &Runner{
		browser:    browser,
		wappalyzer: wap,
		options:    opts,
		writers:    writers,
		Targets:    make(chan string),
	}, nil
}

// witness does the work of probing a url.
// This is where everything comes together as far as the runner is concerned.
func (run *Runner) witness(target string) {
	logger := log.With("target", target)
	logger.Debug("witnessing 👀")

	page, err := run.browser.Page(proto.TargetCreateTarget{})
	if err != nil {
		logger.Error("could not get a page", "err", err)
		return
	}
	defer page.Close()

	// configure timeout
	duration := time.Duration(run.options.Scan.Timeout) * time.Second
	page = page.Timeout(duration)

	// set user agent
	if err := page.SetUserAgent(&proto.NetworkSetUserAgentOverride{
		UserAgent: run.options.Chrome.UserAgent,
	}); err != nil {
		logger.Error("unable to set user-agent string", "err", err)
		return
	}

	// set extra headers, if any
	if len(run.options.Chrome.Headers) > 0 {
		for _, header := range run.options.Chrome.Headers {
			kv := strings.SplitN(header, ":", 2)
			if len(kv) != 2 {
				logger.Warn("custom header did not parse correctly", "header", header)
				continue
			}

			_, err := page.SetExtraHeaders([]string{kv[0], kv[1]})
			if err != nil {
				logger.Error("could not set extra headers for page", "err", err)
				return
			}
		}
	}

	// use page events to grab information about targets. It's how we
	// know what the results of the first request is to save as an overall
	// url result for output writers.
	var (
		first  = ""
		result = &models.Result{
			URL: target,
		}
		netlog = make(map[string]models.NetworkLog)
	)

	go page.EachEvent(
		// dismiss any javascript dialogs
		func(e *proto.PageJavascriptDialogOpening) {
			_ = proto.PageHandleJavaScriptDialog{Accept: true}.Call(page)
		},

		// log console.* calls
		func(e *proto.RuntimeConsoleAPICalled) {
			v := ""
			for _, arg := range e.Args {
				if !arg.Value.Nil() {
					v += arg.Value.String()
				}
			}

			if v == "" {
				return
			}

			result.Console = append(result.Console, models.ConsoleLog{
				Type:  string(e.Type),
				Value: strings.TrimSpace(v),
			})
		},

		// network related events
		// write a request to the network request map
		func(e *proto.NetworkRequestWillBeSent) {
			// note the request id for the first request. well get back
			// to this afterwards to extract information about the probe.
			if first == "" {
				first = string(e.RequestID)
			}

			// record the new request
			netlog[string(e.RequestID)] = models.NetworkLog{
				Time:        e.WallTime.Time(),
				RequestType: models.HTTP,
				URL:         e.Request.URL,
			}
		},

		// write the response to the network request map
		func(e *proto.NetworkResponseReceived) {
			// grab an existing requestid, and add response info
			if entry, ok := netlog[string(e.RequestID)]; ok {
				// update the first request details
				if first == string(e.RequestID) {
					result.FinalURL = e.Response.URL
					result.ResponseCode = e.Response.Status
					result.ResponseReason = e.Response.StatusText
					result.Protocol = e.Response.Protocol
					result.ContentLength = int64(e.Response.EncodedDataLength)

					// write headers
					for k, v := range e.Response.Headers {
						result.Headers = append(result.Headers, models.Header{
							Key:   k,
							Value: v.Str(),
						})
					}

					// grab security detail if available
					if e.Response.SecurityDetails != nil {
						var sanlist []models.TLSSanList
						for _, san := range e.Response.SecurityDetails.SanList {
							sanlist = append(sanlist, models.TLSSanList{
								Value: san,
							})
						}

						result.TLS = models.TLS{
							Protocol:                 e.Response.SecurityDetails.Protocol,
							KeyExchange:              e.Response.SecurityDetails.KeyExchange,
							Cipher:                   e.Response.SecurityDetails.Cipher,
							SubjectName:              e.Response.SecurityDetails.SubjectName,
							SanList:                  sanlist,
							Issuer:                   e.Response.SecurityDetails.Issuer,
							ValidFrom:                float64(e.Response.SecurityDetails.ValidFrom),
							ValidTo:                  float64(e.Response.SecurityDetails.ValidTo),
							ServerSignatureAlgorithm: e.Response.SecurityDetails.ServerSignatureAlgorithm,
							EncryptedClientHello:     e.Response.SecurityDetails.EncryptedClientHello,
						}
					}
				} else { // else, add a network log
					entry.StatusCode = e.Response.Status
					entry.URL = e.Response.URL
					entry.RemoteIP = e.Response.RemoteIPAddress
					entry.MIMEType = e.Response.MIMEType
					entry.Time = e.Response.ResponseTime.Time()

					// write the network log
					result.Network = append(result.Network, entry)
				}
			}
		},

		// mark a request as failed
		func(e *proto.NetworkLoadingFailed) {
			// grab an existing requestid an add failure info
			if entry, ok := netlog[string(e.RequestID)]; ok {
				// update the first request details
				if first == string(e.RequestID) {
					result.Failed = true
					result.FailedReason = e.ErrorText
				} else {
					entry.Error = e.ErrorText

					// write the network log
					result.Network = append(result.Network, entry)
				}
			}
		},

		// TODO: wss
	)()

	// finally, navigate to the target
	if err := page.Navigate(target); err != nil {
		if run.options.Logging.LogScanErrors {
			logger.Error("could not navigate to target", "err", err)
		}
		return
	}

	// wait for the --delay number of seconds
	page.Timeout(time.Duration(run.options.Scan.Delay) * time.Second).WaitNavigation(proto.PageLifecycleEventNameFirstMeaningfulPaint)()

	// wait for navigation to complete
	if err := page.WaitLoad(); err != nil {
		if run.options.Logging.LogScanErrors {
			logger.Error("could not wait for window.onload", "err", err)
		}
		return
	}

	// sanity check
	if first == "" {
		logger.Error("🤔 could not determine first request. how??")
		return
	}

	// if run any JavaScript if we have
	if run.options.Scan.JavaScript != "" {
		_, err := page.Eval(run.options.Scan.JavaScript)
		if err != nil {
			log.Warn("failed to evaluate user-provided javascript", "err", err)
		}
	}

	// get and set the last results info before triggering the
	info, err := page.Info()
	if err != nil {
		logger.Error("could not get page info", "err", err)
	} else {
		result.Title = info.Title
	}

	html, err := page.HTML()
	if err != nil {
		logger.Error("could not get page html", "err", err)
	} else {
		result.HTML = html
	}

	// fingerprint technologies in the first response
	if fingerprints := run.wappalyzer.Fingerprint(result.HeaderMap(), []byte(result.HTML)); fingerprints != nil {
		for tech := range fingerprints {
			result.Technologies = append(result.Technologies, models.Technology{
				Value: tech,
			})
		}
	}

	// take the screenshot
	logger.Debug("taking a screenshot 🔎")
	var screenshotOptions = &proto.PageCaptureScreenshot{OptimizeForSpeed: true}
	switch run.options.Scan.ScreenshotFormat {
	case "jpeg":
		screenshotOptions.Format = proto.PageCaptureScreenshotFormatJpeg
		screenshotOptions.Quality = gson.Int(90)
	case "png":
		screenshotOptions.Format = proto.PageCaptureScreenshotFormatPng
	}

	img, err := page.Screenshot(run.options.Scan.ScreenshotFullPage, screenshotOptions)
	if err != nil {
		if run.options.Logging.LogScanErrors {
			logger.Error("could not take screenshot", "err", err)
		}
		return
	}

	// write the screenshot to disk
	result.Filename = islazy.SafeFileName(target) + "." + run.options.Scan.ScreenshotFormat
	if err := os.WriteFile(
		filepath.Join(run.options.Scan.ScreenshotPath, result.Filename),
		img, 0o664,
	); err != nil {
		if run.options.Logging.LogScanErrors {
			logger.Error("could not write screenshot to disk", "err", err)
		}
		return
	}

	// calculate and set the perception hash
	decoded, _, err := image.Decode(bytes.NewReader(img))
	if err != nil {
		logger.Error("failed to decode screenshot image", "err", err)
		return
	}

	hash, err := goimagehash.PerceptionHash(decoded)
	if err != nil {
		logger.Error("failed to calculate image perception hash", "err", err)
		return
	}
	result.PerceptionHash = hash.ToString()

	// we have everything we can enumerate!
	// pass the result off the configured writers.
	if err := run.callWriters(result); err != nil {
		logger.Error("failed to write results", "err", err)
	}

	logger.Info("page result", "title", info.Title)
}

// callWriters takes a result and passes it to writers
func (run *Runner) callWriters(result *models.Result) error {
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
				// validate the target
				if err := run.checkUrl(target); err != nil {
					if run.options.Logging.LogScanErrors {
						log.Error("invalid target to scan", "target", target, "err", err)
					}
					continue
				}

				// process the target
				// TODO: bubble an error up from witness()
				run.witness(target)
			}
		}()
	}

	wg.Wait()
}

// Close cleans up the Browser runner. The caller needs
// to close the Targets channel
func (run *Runner) Close() {
	log.Debug("closing this browser instance")

	run.browser.Close()
}
