package driver

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/corona10/goimagehash"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/sensepost/gowitness/internal/islazy"
	"github.com/sensepost/gowitness/pkg/log"
	"github.com/sensepost/gowitness/pkg/models"
	"github.com/sensepost/gowitness/pkg/runner"
	"github.com/ysmood/gson"
)

// Gorod is a driver that probes web targets using go-rod
type Gorod struct {
	// browser is a go-rod browser instance
	browser *rod.Browser
	// user data directory
	userData string
	// options for the Runner to consider
	options runner.Options
	// logger
	log *slog.Logger
}

// New gets a new Runner ready for probing.
// It's up to the caller to call Close() on the instance.
func NewGorod(logger *slog.Logger, opts runner.Options) (*Gorod, error) {
	var (
		url      string
		userData string
		err      error
	)

	if opts.Chrome.WSS == "" {
		userData, err = os.MkdirTemp("", "gowitness-v3-gorod-*")
		if err != nil {
			return nil, err
		}

		// get chrome ready
		chrmLauncher := launcher.New().
			// https://github.com/GoogleChrome/chrome-launcher/blob/main/docs/chrome-flags-for-tools.md
			Set("user-data-dir", userData).
			Set("disable-features", "MediaRouter").
			Set("disable-client-side-phishing-detection").
			Set("explicitly-allowed-ports", restrictedPorts()).
			Set("disable-default-apps").
			Set("hide-scrollbars").
			Set("mute-audio").
			Set("no-default-browser-check").
			Set("no-first-run").
			Set("deny-permission-prompts")

		log.Debug("go-rod chrome args", "args", chrmLauncher.FormatArgs())

		// user specified Chrome
		if opts.Chrome.Path != "" {
			chrmLauncher.Bin(opts.Chrome.Path)
		}

		// proxy
		if opts.Chrome.Proxy != "" {
			chrmLauncher.Proxy(opts.Chrome.Proxy)
		}

		url, err = chrmLauncher.Launch()
		if err != nil {
			return nil, err
		}
		logger.Debug("got a browser up", "control-url", url)
	} else {
		url = opts.Chrome.WSS
		logger.Debug("using a user specified WSS url", "control-url", url)
	}

	// connect to the control-url
	browser := rod.New().ControlURL(url)
	if err := browser.Connect(); err != nil {
		return nil, err
	}

	// ignore cert errors
	if err := browser.IgnoreCertErrors(true); err != nil {
		return nil, err
	}

	return &Gorod{
		browser:  browser,
		userData: userData,
		options:  opts,
		log:      logger,
	}, nil
}

// witness does the work of probing a url.
// This is where everything comes together as far as the runner is concerned.
func (run *Gorod) Witness(target string, runner *runner.Runner) (*models.Result, error) {
	logger := run.log.With("target", target)
	logger.Debug("witnessing 👀")

	page, err := run.browser.Page(proto.TargetCreateTarget{})
	if err != nil {
		return nil, fmt.Errorf("could not get a page: %w", err)
	}
	defer page.Close()

	// configure viewport size
	if run.options.Chrome.WindowX > 0 && run.options.Chrome.WindowY > 0 {
		if err := page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
			Width:  run.options.Chrome.WindowX,
			Height: run.options.Chrome.WindowY,
		}); err != nil {
			return nil, fmt.Errorf("unable to set viewport: %w", err)
		}
	}

	// configure timeout
	duration := time.Duration(run.options.Scan.Timeout) * time.Second
	page = page.Timeout(duration)

	// set user agent
	if err := page.SetUserAgent(&proto.NetworkSetUserAgentOverride{
		UserAgent: run.options.Chrome.UserAgent,
	}); err != nil {
		return nil, fmt.Errorf("unable to set user-agent string: %w", err)
	}

	// set extra headers, if any
	if len(run.options.Chrome.Headers) > 0 {
		var headers []string
		for _, header := range run.options.Chrome.Headers {
			kv := strings.SplitN(header, ":", 2)
			if len(kv) != 2 {
				logger.Warn("custom header did not parse correctly", "header", header)
				continue
			}

			headers = append(headers, strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1]))
		}
		_, err := page.SetExtraHeaders(headers)
		if err != nil {
			return nil, fmt.Errorf("could not set extra headers for page: %s", err)
		}
	}

	// use page events to grab information about targets. It's how we
	// know what the results of the first request is to save as an overall
	// url result for output writers.
	var (
		first  *proto.NetworkRequestWillBeSent
		result = &models.Result{
			URL:      target,
			ProbedAt: time.Now(),
		}
		resultMutex   = sync.Mutex{}
		netlog        = make(map[string]models.NetworkLog)
		dismissEvents = false // set to true to stop EachEvent callbacks
	)

	go page.EachEvent(
		// dismiss any javascript dialogs
		func(e *proto.PageJavascriptDialogOpening) bool {
			_ = proto.PageHandleJavaScriptDialog{Accept: true}.Call(page)
			return dismissEvents
		},

		// log console.* calls
		func(e *proto.RuntimeConsoleAPICalled) bool {
			v := ""
			for _, arg := range e.Args {
				if !arg.Value.Nil() {
					v += arg.Value.String()
				}
			}

			if v == "" {
				return dismissEvents
			}

			resultMutex.Lock()
			result.Console = append(result.Console, models.ConsoleLog{
				Type:  "console." + string(e.Type),
				Value: strings.TrimSpace(v),
			})
			resultMutex.Unlock()

			return dismissEvents
		},

		// network related events
		// write a request to the network request map
		func(e *proto.NetworkRequestWillBeSent) bool {
			// note the request id for the first request. well get back
			// to this afterwards to extract information about the probe.
			if first == nil {
				first = e
			}

			// record the new request
			netlog[string(e.RequestID)] = models.NetworkLog{
				Time:        e.WallTime.Time(),
				RequestType: models.HTTP,
				URL:         e.Request.URL,
			}

			return dismissEvents
		},

		// write the response to the network request map
		func(e *proto.NetworkResponseReceived) bool {
			// grab an existing requestid, and add response info
			if entry, ok := netlog[string(e.RequestID)]; ok {
				// update the first request details (headers, tls, etc.)
				if first != nil && first.RequestID == e.RequestID {
					resultMutex.Lock()
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
							ValidFrom:                islazy.Float64ToTime(float64(e.Response.SecurityDetails.ValidFrom)),
							ValidTo:                  islazy.Float64ToTime(float64(e.Response.SecurityDetails.ValidTo)),
							ServerSignatureAlgorithm: int64(*e.Response.SecurityDetails.ServerSignatureAlgorithm),
							EncryptedClientHello:     e.Response.SecurityDetails.EncryptedClientHello,
						}
					}
					resultMutex.Unlock()
				}

				entry.StatusCode = int64(e.Response.Status)
				entry.URL = e.Response.URL
				entry.RemoteIP = e.Response.RemoteIPAddress
				entry.MIMEType = e.Response.MIMEType
				entry.Time = e.Response.ResponseTime.Time()

				// write the network log
				resultMutex.Lock()
				entryIndex := len(result.Network)
				result.Network = append(result.Network, entry)
				resultMutex.Unlock()

				// if we need to write the body, do that
				if run.options.Scan.SaveContent {
					go func(index int) {
						body, err := proto.NetworkGetResponseBody{RequestID: e.RequestID}.Call(page)
						if err != nil {
							if run.options.Logging.LogScanErrors {
								if run.options.Logging.LogScanErrors {
									run.log.Error("could not get network request response body", "url", e.Response.URL, "err", err)
								}
								return
							}
						}

						resultMutex.Lock()
						result.Network[index].Content = []byte(body.Body)
						resultMutex.Unlock()
					}(entryIndex)
				}
			}

			return dismissEvents
		},

		// mark a request as failed
		func(e *proto.NetworkLoadingFailed) bool {
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

			return dismissEvents
		},

		// TODO: wss
	)()

	// finally, navigate to the target
	if err := page.Navigate(target); err != nil {
		return nil, fmt.Errorf("could not navigate to target: %s", err)
	}

	// wait for the configured delay
	if run.options.Scan.Delay > 0 {
		time.Sleep(time.Duration(run.options.Scan.Delay) * time.Second)
	}

	// run any javascript we have
	if run.options.Scan.JavaScript != "" {
		_, err := page.Eval(run.options.Scan.JavaScript)
		if err != nil {
			logger.Warn("failed to evaluate user-provided javascript", "err", err)
		}
	}

	// get cookies
	cookies, err := page.Cookies([]string{})
	if err != nil {
		if run.options.Logging.LogScanErrors {
			logger.Error("could not get cookies", "err", err)
		}
	} else {
		for _, cookie := range cookies {
			result.Cookies = append(result.Cookies, models.Cookie{
				Name:         cookie.Name,
				Value:        cookie.Value,
				Domain:       cookie.Domain,
				Path:         cookie.Path,
				Expires:      cookie.Expires.Time(),
				Size:         int64(cookie.Size),
				HTTPOnly:     cookie.HTTPOnly,
				Secure:       cookie.Secure,
				Session:      cookie.Session,
				Priority:     string(cookie.Priority),
				SourceScheme: string(cookie.SourceScheme),
				SourcePort:   int64(cookie.SourcePort),
			})
		}
	}

	// get and set the last results info before triggering the
	info, err := page.Info()
	if err != nil {
		if run.options.Logging.LogScanErrors {
			logger.Error("could not get page info", "err", err)
		}
	} else {
		result.Title = info.Title
	}

	if !run.options.Scan.SkipHTML {
		html, err := page.HTML()
		if err != nil {
			if run.options.Logging.LogScanErrors {
				logger.Error("could not get page html", "err", err)
			}
		} else {
			result.HTML = html
		}
	}

	// stop the event handlers
	dismissEvents = true

	// fingerprint technologies in the first response
	if fingerprints := runner.Wappalyzer.Fingerprint(result.HeaderMap(), []byte(result.HTML)); fingerprints != nil {
		for tech := range fingerprints {
			result.Technologies = append(result.Technologies, models.Technology{
				Value: tech,
			})
		}
	}

	// take the screenshot. getting here often means the page responded and we have
	// some information. sometimes though, and im not sure why, page.Screenshot()
	// fails by timing out. in that case, record what we have at least but martk
	// the screenshotting as failed. that way we dont lose all our work at least.
	logger.Debug("taking a screenshot 🔎")
	var screenshotOptions = &proto.PageCaptureScreenshot{}
	switch run.options.Scan.ScreenshotFormat {
	case "jpeg":
		screenshotOptions.Format = proto.PageCaptureScreenshotFormatJpeg
		screenshotOptions.Quality = gson.Int(80)
	case "png":
		screenshotOptions.Format = proto.PageCaptureScreenshotFormatPng
	}

	img, err := page.Screenshot(run.options.Scan.ScreenshotFullPage, screenshotOptions)
	if err != nil {
		if run.options.Logging.LogScanErrors {
			logger.Error("could not grab screenshot", "err", err)
		}

		result.Failed = true
		result.FailedReason = err.Error()
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

		hash, err := goimagehash.PerceptionHash(decoded)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate image perception hash: %w", err)
		}
		result.PerceptionHash = hash.ToString()
	}

	return result, nil
}

// Close cleans up the Browser runner. The caller needs
// to close the Targets channel
func (run *Gorod) Close() {
	run.log.Debug("closing the browser instance")

	if err := run.browser.Close(); err != nil {
		log.Error("could not close the browser", "err", err)
		return
	}

	// cleaning user data
	if run.userData != "" {
		// wait a sec for the browser process to go away
		time.Sleep(time.Second * 1)

		run.log.Debug("cleaning user data directory", "directory", run.userData)
		if err := os.RemoveAll(run.userData); err != nil {
			run.log.Error("could not cleanup temporary user data dir", "dir", run.userData, "err", err)
		}
	}
}
