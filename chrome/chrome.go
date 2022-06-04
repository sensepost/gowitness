package chrome

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/chromedp/cdproto/inspector"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/sensepost/gowitness/storage"
	"gorm.io/gorm"
)

// Chrome contains information about a Google Chrome
// instance, with methods to run on it.
type Chrome struct {
	ResolutionX int
	ResolutionY int
	UserAgent   string
	Timeout     int64
	Delay       int
	FullPage    bool
	ChromePath  string
	Proxy       string
	Headers     []string
	HeadersMap  map[string]interface{}

	// wappalyzer client
	wappalyzer *Wappalyzer
}

// PreflightResult contains the results of a preflight run
type PreflightResult struct {
	URL              *url.URL
	HTTPResponse     *http.Response
	HTTPTitle        string
	HTTPTechnologies []string
}

// NewChrome returns a new initialised Chrome struct
func NewChrome() *Chrome {
	return &Chrome{
		wappalyzer: NewWappalyzer(),
	}
}

// Preflight will preflight a url
func (chrome *Chrome) Preflight(url *url.URL) (result *PreflightResult, err error) {

	// init a new preflight result
	result = &PreflightResult{}

	// purposefully ignore bad certs
	transport := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives: true,
	}

	if chrome.Proxy != "" {
		var erri error
		proxyURL, erri := url.Parse(chrome.Proxy)
		if erri != nil {
			return
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	// purposefully ignore bad certs
	client := http.Client{
		Transport: transport,
	}

	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return
	}
	req.Header.Set("User-Agent", chrome.UserAgent)

	// set the preflight headers (type assertion for value)
	for k, v := range chrome.HeadersMap {
		req.Header.Set(k, v.(string))
	}

	req.Close = true

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(chrome.Timeout)*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	result.URL = url
	result.HTTPResponse = resp

	// if we cant perform wappalyzer lookups, then return
	if chrome.wappalyzer.err != nil {
		return
	}

	result.HTTPTitle = chrome.wappalyzer.HTMLTitle(body)
	result.HTTPTechnologies = chrome.wappalyzer.Technologies(req.Header, body)

	return
}

// StorePreflight will store preflight info to a DB
func (chrome *Chrome) StorePreflight(db *gorm.DB, preflight *PreflightResult, filename string) (uint, error) {

	record := &storage.URL{
		URL:            preflight.URL.String(),
		FinalURL:       preflight.HTTPResponse.Request.URL.String(),
		ResponseCode:   preflight.HTTPResponse.StatusCode,
		ResponseReason: preflight.HTTPResponse.Status,
		Proto:          preflight.HTTPResponse.Proto,
		ContentLength:  preflight.HTTPResponse.ContentLength,
		Title:          preflight.HTTPTitle,
		Filename:       filename,
	}

	// append headers
	for k, v := range preflight.HTTPResponse.Header {
		hv := strings.Join(v, ", ")
		record.AddHeader(k, hv)
	}

	for _, v := range preflight.HTTPTechnologies {
		record.AddTechnologie(v)
	}

	// get TLS info, if any
	if preflight.HTTPResponse.TLS != nil {
		record.TLS = storage.TLS{
			Version:    preflight.HTTPResponse.TLS.Version,
			ServerName: preflight.HTTPResponse.TLS.ServerName,
		}

		for _, cert := range preflight.HTTPResponse.TLS.PeerCertificates {
			tlsCert := &storage.TLSCertificate{
				SubjectCommonName:  cert.Subject.CommonName,
				IssuerCommonName:   cert.Issuer.CommonName,
				SignatureAlgorithm: cert.SignatureAlgorithm.String(),
				PubkeyAlgorithm:    cert.PublicKeyAlgorithm.String(),
			}

			for _, name := range cert.DNSNames {
				tlsCert.AddDNSName(name)
			}

			record.TLS.TLSCertificates = append(record.TLS.TLSCertificates, *tlsCert)
		}
	}

	db.Create(record)
	return record.ID, nil
}

// Screenshot takes a screenshot of a URL and saves it to destination
// Ref:
// 	https://github.com/chromedp/examples/blob/255873ca0d76b00e0af8a951a689df3eb4f224c3/screenshot/main.go
func (chrome *Chrome) Screenshot(url *url.URL) ([]byte, error) {

	// setup chromedp default options
	options := []chromedp.ExecAllocatorOption{}
	options = append(options, chromedp.DefaultExecAllocatorOptions[:]...)
	options = append(options, chromedp.UserAgent(chrome.UserAgent))
	options = append(options, chromedp.DisableGPU)
	options = append(options, chromedp.Flag("ignore-certificate-errors", true)) // RIP shittyproxy.go
	options = append(options, chromedp.WindowSize(chrome.ResolutionX, chrome.ResolutionY))

	if chrome.ChromePath != "" {
		options = append(options, chromedp.ExecPath(chrome.ChromePath))
	}

	if chrome.Proxy != "" {
		options = append(options, chromedp.ProxyServer(chrome.Proxy))
	}

	actx, acancel := chromedp.NewExecAllocator(context.Background(), options...)
	defer acancel()
	browserCtx, cancelBrowserCtx := chromedp.NewContext(actx)
	defer cancelBrowserCtx()

	// create the initial context to act as the 'tab', where we will perform the initial navigation
	// if this context loads successfully, then the screenshot will have been captured
	//
	//		Note:	You're not supposed to delay the initial run context, so we use WithTimeout
	//				 https://pkg.go.dev/github.com/chromedp/chromedp#Run

	tabCtx, cancelTabCtx := context.WithTimeout(browserCtx, time.Duration(chrome.Timeout)*time.Second)
	defer cancelTabCtx()

	// Run the initial browser
	if err := chromedp.Run(browserCtx); err != nil {
		return nil, err
	}

	var buf []byte

	// prevent browser crashes from locking the context (prevents hanging)
	chromedp.ListenTarget(browserCtx, func(ev interface{}) {
		if _, ok := ev.(*inspector.EventTargetCrashed); ok {
			cancelBrowserCtx()
		}
	})

	chromedp.ListenTarget(tabCtx, func(ev interface{}) {
		if _, ok := ev.(*inspector.EventTargetCrashed); ok {
			cancelTabCtx()
		}
	})

	// squash JavaScript dialog boxes such as alert();
	chromedp.ListenTarget(tabCtx, func(ev interface{}) {
		if _, ok := ev.(*page.EventJavascriptDialogOpening); ok {
			go func() {
				if err := chromedp.Run(tabCtx,
					page.HandleJavaScriptDialog(true),
				); err != nil {
					cancelTabCtx()
				}
			}()
		}
	})

	// perform navigation on the tab context and attempt to take a clean screenshot
	err := chromedp.Run(tabCtx, buildTasks(chrome, url, true, &buf))

	if errors.Is(err, context.DeadlineExceeded) {
		// if the context timeout exceeded (e.g. on a long page load) then
		// just take the screenshot this will take a screenshot of whatever
		// loaded before failing

		// create a new tab context for this scenario, since our previous
		// context expired using a context timeout delay again to help
		// prevent hanging scenarios
		newTabCtx, cancelNewTabCtx := context.WithTimeout(browserCtx, time.Duration(chrome.Timeout)*time.Second)
		defer cancelNewTabCtx()

		// listen for crashes on this backup context as well
		chromedp.ListenTarget(newTabCtx, func(ev interface{}) {
			if _, ok := ev.(*inspector.EventTargetCrashed); ok {
				cancelNewTabCtx()
			}
		})

		// attempt to capture the screenshot of the tab and replace error accordingly
		err = chromedp.Run(newTabCtx, buildTasks(chrome, url, false, &buf))
	}

	if err != nil {
		return nil, err
	}

	return buf, nil
}

// builds the chromedp tasks slice in a bit cleaner fashion with the various logic
func buildTasks(chrome *Chrome, url *url.URL, doNavigate bool, buf *[]byte) chromedp.Tasks {
	var actions chromedp.Tasks

	if len(chrome.HeadersMap) > 0 {
		actions = append(actions, network.Enable(), network.SetExtraHTTPHeaders(network.Headers(chrome.HeadersMap)))
	}

	if doNavigate {
		actions = append(actions, chromedp.Navigate(url.String()))
		if chrome.Delay > 0 {
			actions = append(actions, chromedp.Sleep(time.Duration(chrome.Delay)*time.Second))
		}
		actions = append(actions, chromedp.Stop())
	}

	if chrome.FullPage {
		actions = append(actions, chromedp.FullScreenshot(buf, 100))
	} else {
		actions = append(actions, chromedp.CaptureScreenshot(buf))
	}

	return actions
}

// initalize the headers Map. we do this given the format chromedp wants
// Ref:
// 	https://github.com/chromedp/examples/blob/master/headers/main.go
func (chrome *Chrome) PrepareHeaderMap() {

	if len(chrome.Headers) <= 0 {
		return
	}

	// initialize the map
	chrome.HeadersMap = make(map[string]interface{})

	// split each header string and append to the map
	for _, header := range chrome.Headers {

		headerSlice := strings.SplitN(header, ":", 2)
		// add header to the map
		if len(headerSlice) == 2 {
			chrome.HeadersMap[headerSlice[0]] = headerSlice[1]
		}
	}
}
