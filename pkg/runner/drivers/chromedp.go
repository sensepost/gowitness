package driver

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/cdproto/storage"
	targetproto "github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
	"github.com/sensepost/gowitness/internal/islazy"
	"github.com/sensepost/gowitness/pkg/imagehash"
	"github.com/sensepost/gowitness/pkg/models"
	"github.com/sensepost/gowitness/pkg/runner"
)

// Chromedp is a driver that probes web targets using chromedp
// Implementation ref: https://github.com/chromedp/examples/blob/master/multi/main.go
type Chromedp struct {
	// options for the Runner to consider
	options runner.Options

	// logger
	log *slog.Logger

	// allocator and browser context shared for tabs
	allocator     *browserInstance
	browserCtx    context.Context
	browserCancel context.CancelFunc

	// shared task template that will be cloned per witness
	baseTasks chromedp.Tasks
}

// browserInstance is an instance used by one run of Witness
type browserInstance struct {
	allocCtx    context.Context
	allocCancel context.CancelFunc
	userData    string
}

// Close closes the allocator, and cleans up the user dir.
func (b *browserInstance) Close() {
	b.allocCancel()
	<-b.allocCtx.Done()

	// cleanup the user data directory
	if b.userData != "" {
		os.RemoveAll(b.userData)
	}
}

// getChromedpAllocator is a helper function to get a chrome allocation context.
func getChromedpAllocator(opts runner.Options) (*browserInstance, error) {
	var (
		allocCtx    context.Context
		allocCancel context.CancelFunc
		userData    string
		err         error
	)

	if opts.Chrome.WSS == "" {
		userData, err = os.MkdirTemp("", "gowitness-v3-chromedp-*")
		if err != nil {
			return nil, err
		}

		// set up chrome context and launch options
		allocOpts := append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.NoDefaultBrowserCheck,
			chromedp.NoFirstRun,
			chromedp.DisableGPU,
			chromedp.IgnoreCertErrors,
			chromedp.UserAgent(opts.Chrome.UserAgent),
			chromedp.Flag("disable-features", "MediaRouter"),
			chromedp.Flag("mute-audio", true),
			chromedp.Flag("hide-scrollbars", true),
			chromedp.Flag("disable-background-timer-throttling", true),
			chromedp.Flag("disable-backgrounding-occluded-windows", true),
			chromedp.Flag("disable-renderer-backgrounding", true),
			chromedp.Flag("deny-permission-prompts", true),
			chromedp.Flag("https-upgrades-enabled", false),
			chromedp.Flag("disable-features", "HttpsUpgrades"),
			chromedp.Flag("explicitly-allowed-ports", restrictedPorts()),
			chromedp.Flag("no-sandbox", true),
			chromedp.WindowSize(opts.Chrome.WindowX, opts.Chrome.WindowY),
			chromedp.UserDataDir(userData),
		)

		// Set proxy if specified
		if opts.Chrome.Proxy != "" {
			allocOpts = append(allocOpts, chromedp.ProxyServer(opts.Chrome.Proxy))
		}

		// Use specific Chrome binary if provided
		if opts.Chrome.Path != "" {
			allocOpts = append(allocOpts, chromedp.ExecPath(opts.Chrome.Path))
		}

		allocCtx, allocCancel = chromedp.NewExecAllocator(context.Background(), allocOpts...)

	} else {
		allocCtx, allocCancel = chromedp.NewRemoteAllocator(context.Background(), opts.Chrome.WSS)
	}

	return &browserInstance{
		allocCtx:    allocCtx,
		allocCancel: allocCancel,
		userData:    userData,
	}, nil
}

// NewChromedp returns a new Chromedp instance
func NewChromedp(logger *slog.Logger, opts runner.Options) (*Chromedp, error) {
	allocator, err := getChromedpAllocator(opts)
	if err != nil {
		return nil, err
	}

	browserCtx, browserCancel := chromedp.NewContext(allocator.allocCtx)

	// warm up chrome so that the singleton lock is held before worker goroutines start
	if err := chromedp.Run(browserCtx, chromedp.ActionFunc(func(context.Context) error { return nil })); err != nil {
		browserCancel()
		allocator.Close()
		return nil, fmt.Errorf("failed to initialize chrome context: %w", err)
	}

	driver := &Chromedp{
		options:       opts,
		log:           logger,
		allocator:     allocator,
		browserCtx:    browserCtx,
		browserCancel: browserCancel,
	}

	driver.baseTasks = chromedp.Tasks{
		network.Enable(),
	}

	return driver, nil
}

// witness does the work of probing a url.
// This is where everything comes together as far as the runner is concerned.
func (run *Chromedp) Witness(target string, thisRunner *runner.Runner) (*models.Result, error) {
	logger := run.log.With("target", target)
	logger.Debug("witnessing 👀")

	// get a tab
	tabCtx, tabCancel := chromedp.NewContext(run.browserCtx)
	tabChromedpCtx := chromedp.FromContext(tabCtx)

	// get the tab ID, so we can close it when we are done
	var tabID targetproto.ID
	if tabChromedpCtx != nil && tabChromedpCtx.Target != nil {
		tabID = tabChromedpCtx.Target.TargetID
	}

	defer func() {
		if tabID != "" {
			closeCtx, cancel := context.WithTimeout(run.browserCtx, 5*time.Second)
			defer cancel()
			if err := chromedp.Run(closeCtx, chromedp.ActionFunc(func(ctx context.Context) error {
				return targetproto.CloseTarget(tabID).Do(ctx)
			})); err != nil && run.options.Logging.LogScanErrors {
				logger.Error("failed to close tab", "err", err)
			}
		}

		tabCancel()
	}()

	// get a timeout context for navigation
	navigationCtx, navigationCancel := context.WithTimeout(tabCtx, time.Duration(run.options.Scan.Timeout)*time.Second)
	defer navigationCancel()

	// use page events to grab information about targets. It's how we
	// know what the results of the first request is to save as an overall
	// url result for output writers.
	var (
		result = &models.Result{
			URL:      target,
			ProbedAt: time.Now(),
		}
		resultMutex sync.Mutex
		first       *network.EventRequestWillBeSent
		netlog      = make(map[string]models.NetworkLog)
	)

	go chromedp.ListenTarget(navigationCtx, func(ev interface{}) {
		switch e := ev.(type) {
		// dismiss any javascript dialogs
		case *page.EventJavascriptDialogOpening:
			if ctx := chromedp.FromContext(navigationCtx); ctx != nil && ctx.Target != nil {
				if err := page.HandleJavaScriptDialog(true).Do(cdp.WithExecutor(navigationCtx, ctx.Target)); err != nil {
					logger.Error("failed to handle a javascript dialog", "err", err)
				}
			}
		// log console.* calls
		case *runtime.EventConsoleAPICalled:
			v := ""
			for _, arg := range e.Args {
				v += string(arg.Value)
			}

			if v == "" {
				return
			}

			resultMutex.Lock()
			result.Console = append(result.Console, models.ConsoleLog{
				Type:  "console." + string(e.Type),
				Value: strings.TrimSpace(v),
			})
			resultMutex.Unlock()

		// network related events
		// write a request to the network request map
		case *network.EventRequestWillBeSent:
			if first == nil {
				first = e
			}
			netlog[string(e.RequestID)] = models.NetworkLog{
				Time:        e.WallTime.Time(),
				RequestType: models.HTTP,
				URL:         e.Request.URL,
			}
		case *network.EventResponseReceived:
			if entry, ok := netlog[string(e.RequestID)]; ok {
				if first != nil && first.RequestID == e.RequestID {
					resultMutex.Lock()
					result.FinalURL = e.Response.URL
					result.ResponseCode = int(e.Response.Status)
					result.ResponseReason = e.Response.StatusText
					result.Protocol = e.Response.Protocol
					result.ContentLength = int64(e.Response.EncodedDataLength)

					// write headers
					for k, v := range e.Response.Headers {
						result.Headers = append(result.Headers, models.Header{
							Key:   k,
							Value: v.(string),
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

						// urgh, paaaaain.
						var validFromTime, validToTime time.Time
						if e.Response.SecurityDetails.ValidFrom != nil {
							validFromTime = e.Response.SecurityDetails.ValidFrom.Time()
						}
						if e.Response.SecurityDetails.ValidTo != nil {
							validToTime = e.Response.SecurityDetails.ValidTo.Time()
						}

						result.TLS = models.TLS{
							Protocol:                 e.Response.SecurityDetails.Protocol,
							KeyExchange:              e.Response.SecurityDetails.KeyExchange,
							Cipher:                   e.Response.SecurityDetails.Cipher,
							SubjectName:              e.Response.SecurityDetails.SubjectName,
							SanList:                  sanlist,
							Issuer:                   e.Response.SecurityDetails.Issuer,
							ValidFrom:                validFromTime,
							ValidTo:                  validToTime,
							ServerSignatureAlgorithm: e.Response.SecurityDetails.ServerSignatureAlgorithm,
							EncryptedClientHello:     e.Response.SecurityDetails.EncryptedClientHello,
						}
					}
					resultMutex.Unlock()
				}

				entry.StatusCode = e.Response.Status
				entry.URL = e.Response.URL
				entry.RemoteIP = e.Response.RemoteIPAddress
				entry.MIMEType = e.Response.MimeType
				if e.Response.ResponseTime != nil {
					entry.Time = e.Response.ResponseTime.Time()
				}

				// write the network log
				resultMutex.Lock()
				entryIndex := len(result.Network)
				result.Network = append(result.Network, entry)
				resultMutex.Unlock()

				// if we need to write the body, do that
				// https://github.com/chromedp/chromedp/issues/543
				if run.options.Scan.SaveContent {
					go func(index int) {
						c := chromedp.FromContext(navigationCtx)
						p := network.GetResponseBody(e.RequestID)
						body, err := p.Do(cdp.WithExecutor(navigationCtx, c.Target))
						if err != nil {
							if run.options.Logging.LogScanErrors {
								run.log.Error("could not get network request response body", "url", e.Response.URL, "err", err)
								return
							}
						}

						resultMutex.Lock()
						result.Network[index].Content = body
						resultMutex.Unlock()

					}(entryIndex)
				}
			}
		// mark a request as failed
		case *network.EventLoadingFailed:
			// grab an existing requestid an add failure info
			if entry, ok := netlog[string(e.RequestID)]; ok {
				resultMutex.Lock()

				// update the first request details
				if first != nil && first.RequestID == e.RequestID {
					result.Failed = true
					result.FailedReason = e.ErrorText
				} else {
					entry.Error = e.ErrorText

					// write the network log
					result.Network = append(result.Network, entry)
				}

				resultMutex.Unlock()
			}
		}

		// TODO: wss
	})

	// get cookies
	var cookies []*network.Cookie

	// grab a screenshot
	var (
		img           []byte
		screenshotErr error
	)

	// start a tasks set
	tasks := append(chromedp.Tasks{}, run.baseTasks...)

	// set extra headers, if any
	if len(run.options.Chrome.Headers) > 0 {
		headers := make(network.Headers)
		for _, header := range run.options.Chrome.Headers {
			kv := strings.SplitN(header, ":", 2)
			if len(kv) != 2 {
				logger.Warn("custom header did not parse correctly", "header", header)
				continue
			}

			headers[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}

		tasks = append(tasks, network.SetExtraHTTPHeaders(headers))
	}

	// accumulate tasks to execute in the tab context.
	tasks = append(tasks,
		chromedp.Navigate(target),
		chromedp.WaitReady("body", chromedp.ByQuery),
	)

	if run.options.Scan.Delay > 0 {
		tasks = append(tasks, chromedp.Sleep(time.Duration(run.options.Scan.Delay)*time.Second))
	}

	if run.options.Scan.JavaScript != "" {
		tasks = append(tasks, chromedp.ActionFunc(func(ctx context.Context) error {
			if err := chromedp.Evaluate(run.options.Scan.JavaScript, nil).Do(ctx); err != nil {
				return fmt.Errorf("failed to evaluate user-provided javascript: %w", err)
			}
			return nil
		}))
	}

	tasks = append(tasks, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		cookies, err = storage.GetCookies().Do(ctx)
		if err != nil && run.options.Logging.LogScanErrors {
			logger.Error("could not get cookies", "err", err)
		}
		return nil
	}))

	tasks = append(tasks, chromedp.ActionFunc(func(ctx context.Context) error {
		if err := chromedp.Title(&result.Title).Do(ctx); err != nil && run.options.Logging.LogScanErrors {
			logger.Error("could not get page title", "err", err)
		}
		return nil
	}))

	if !run.options.Scan.SkipHTML {
		tasks = append(tasks, chromedp.ActionFunc(func(ctx context.Context) error {
			if err := chromedp.OuterHTML(":root", &result.HTML, chromedp.ByQueryAll).Do(ctx); err != nil && run.options.Logging.LogScanErrors {
				logger.Error("could not get page html", "err", err)
			}
			return nil
		}))
	}

	tasks = append(tasks, chromedp.ActionFunc(func(ctx context.Context) error {
		params := page.CaptureScreenshot().
			WithQuality(int64(run.options.Scan.ScreenshotJpegQuality)).
			WithFormat(page.CaptureScreenshotFormat(run.options.Scan.ScreenshotFormat))

		if run.options.Scan.ScreenshotFullPage {
			params = params.WithCaptureBeyondViewport(true)
		}

		var err error
		img, err = params.Do(ctx)
		if err != nil {
			screenshotErr = err
		}

		return nil
	}))

	// run the accumulated tasks
	if err := chromedp.Run(navigationCtx, tasks); err != nil && err != context.DeadlineExceeded {
		// check if the error is chrome not found related, in which case
		// we'll return a special error type.
		var execErr *exec.Error
		if errors.As(err, &execErr) && execErr.Err == exec.ErrNotFound {
			return nil, &runner.ChromeNotFoundError{Err: err}
		}

		return nil, err
	}

	if len(cookies) > 0 {
		for _, cookie := range cookies {
			result.Cookies = append(result.Cookies, models.Cookie{
				Name:         cookie.Name,
				Value:        cookie.Value,
				Domain:       cookie.Domain,
				Path:         cookie.Path,
				Expires:      islazy.Float64ToTime(cookie.Expires),
				Size:         cookie.Size,
				HTTPOnly:     cookie.HTTPOnly,
				Secure:       cookie.Secure,
				Session:      cookie.Session,
				Priority:     cookie.Priority.String(),
				SourceScheme: cookie.SourceScheme.String(),
				SourcePort:   cookie.SourcePort,
			})
		}
	}

	// check if the preflight returned a code to filter
	if (len(run.options.Scan.HttpCodeFilter) > 0) && !islazy.SliceHasInt(run.options.Scan.HttpCodeFilter, result.ResponseCode) {
		logger.Warn("http response code was filtered", "code", result.ResponseCode)

		return nil, fmt.Errorf("http response code was %d which is filtered", result.ResponseCode)
	}

	// fingerprint technologies in the first response
	if fingerprints := thisRunner.Wappalyzer.Fingerprint(result.HeaderMap(), []byte(result.HTML)); fingerprints != nil {
		for tech := range fingerprints {
			result.Technologies = append(result.Technologies, models.Technology{
				Value: tech,
			})
		}
	}

	if img == nil && screenshotErr == nil {
		screenshotErr = errors.New("screenshot not captured")
	}

	if screenshotErr != nil {
		if run.options.Logging.LogScanErrors {
			logger.Error("could not grab screenshot", "err", screenshotErr)
		}

		result.Failed = true
		result.FailedReason = screenshotErr.Error()
	} else {
		// give the writer a screenshot to deal with
		if run.options.Scan.ScreenshotToWriter {
			result.Screenshot = base64.StdEncoding.EncodeToString(img)
		}

		// write the screenshot to disk if we have a path
		if !run.options.Scan.ScreenshotSkipSave {
			result.Filename = islazy.SafeFileName(target) + "." + run.options.Scan.ScreenshotFormat
			result.Filename = islazy.LeftTrucate(result.Filename, 200)
			if err := os.WriteFile(
				filepath.Join(run.options.Scan.ScreenshotPath, result.Filename),
				img, os.FileMode(0664),
			); err != nil {
				return nil, fmt.Errorf("could not write screenshot to disk: %w", err)
			}
		}

		// calculate and set the perception hash
		decoded, _, err := image.Decode(bytes.NewReader(img))
		if err != nil {
			return nil, fmt.Errorf("failed to decode screenshot image: %w", err)
		}

		hash, err := imagehash.PerceptionHash(decoded)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate image perception hash: %w", err)
		}
		result.PerceptionHash = hash
	}

	return result, nil
}

func (run *Chromedp) Close() {
	run.log.Debug("closing browser allocation context")
	if run.browserCancel != nil {
		run.browserCancel()
	}
	if run.allocator != nil {
		run.allocator.Close()
	}
}
