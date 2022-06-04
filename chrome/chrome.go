package chrome

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

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

	// fill in wappalyzer information if we have a client to do so
	if chrome.wappalyzer.err != nil {
		result.HTTPTitle = chrome.wappalyzer.HTMLTitle(body)
		result.HTTPTechnologies = chrome.wappalyzer.Technologies(req.Header, body)
	}

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
	ctx, cancel := chromedp.NewContext(actx)
	defer acancel()
	defer cancel()

	var buf []byte

	// squash JavaScript dialog boxes such as alert();
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		if _, ok := ev.(*page.EventJavascriptDialogOpening); ok {
			go func() {
				if err := chromedp.Run(ctx,
					page.HandleJavaScriptDialog(true),
				); err != nil {
					panic(err)
				}
			}()
		}
	})

	if chrome.FullPage {
		// straight from: https://github.com/chromedp/examples/blob/849108f7da9f743bcdaef449699ed57cb4053379/screenshot/main.go

		// additional headers
		if len(chrome.HeadersMap) > 0 {
			if err := chromedp.Run(ctx, chromedp.Tasks{
				network.Enable(),
				network.SetExtraHTTPHeaders(network.Headers(chrome.HeadersMap)),
				chromedp.Navigate(url.String()),
				chromedp.Sleep(time.Duration(chrome.Delay) * time.Second),
				chromedp.FullScreenshot(&buf, 100),
			}); err != nil {
				return nil, err
			}
		} else {
			if err := chromedp.Run(ctx, chromedp.Tasks{
				chromedp.Navigate(url.String()),
				chromedp.Sleep(time.Duration(chrome.Delay) * time.Second),
				chromedp.FullScreenshot(&buf, 100),
			}); err != nil {
				return nil, err
			}
		}

	} else {
		// normal viewport screenshot

		// additional headers
		if len(chrome.HeadersMap) > 0 {
			if err := chromedp.Run(ctx, chromedp.Tasks{
				network.Enable(),
				network.SetExtraHTTPHeaders(network.Headers(chrome.HeadersMap)),
				chromedp.Navigate(url.String()),
				chromedp.Sleep(time.Duration(chrome.Delay) * time.Second),
				chromedp.CaptureScreenshot(&buf),
			}); err != nil {
				return nil, err
			}
		} else {
			if err := chromedp.Run(ctx, chromedp.Tasks{
				chromedp.Navigate(url.String()),
				chromedp.Sleep(time.Duration(chrome.Delay) * time.Second),
				chromedp.CaptureScreenshot(&buf),
			}); err != nil {
				return nil, err
			}
		}
	}

	return buf, nil
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
