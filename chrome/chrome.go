package chrome

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/inspector"
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
}

// NewChrome returns a new initialised Chrome struct
func NewChrome() *Chrome {
	return &Chrome{}
}

// Preflight will preflight a url
func (chrome *Chrome) Preflight(url *url.URL) (resp *http.Response, title string, technologies []string, err error) {
	// purposefully ignore bad certs
	transport := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives: true,
	}

	if chrome.Proxy != "" {
		var erri error
		proxyURL, erri := url.Parse(chrome.Proxy)
		if erri != nil {
			return nil, "", nil, erri
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
	req.Close = true

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(chrome.Timeout)*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	resp, err = client.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()
	title, _ = GetHTMLTitle(resp.Body)
	technologies, _ = GetTechnologies(resp)

	return
}

// StorePreflight will store preflight info to a DB
func (chrome *Chrome) StorePreflight(url *url.URL, db *gorm.DB, resp *http.Response, title string, technologies []string, filename string) (uint, error) {

	record := &storage.URL{
		URL:            url.String(),
		FinalURL:       resp.Request.URL.String(),
		ResponseCode:   resp.StatusCode,
		ResponseReason: resp.Status,
		Proto:          resp.Proto,
		ContentLength:  resp.ContentLength,
		Filename:       filename,
		Title:          title,
	}

	// append headers
	for k, v := range resp.Header {
		hv := strings.Join(v, ", ")
		record.AddHeader(k, hv)
	}

	for _, v := range technologies {
		record.AddTechnologie(v)
	}

	// get TLS info, if any
	if resp.TLS != nil {
		record.TLS = storage.TLS{
			Version:    resp.TLS.Version,
			ServerName: resp.TLS.ServerName,
		}

		for _, cert := range resp.TLS.PeerCertificates {
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
	browserCtx, cancel := chromedp.NewContext(actx)
	defer acancel()
	defer cancel()

	// create the initial context to act as the 'tab', where we will perform the initial navigation
	// if this context loads successfully, then the screenshot will have been captured
	//
	//		Note:	You're not supposed to delay the initial run context, so we use WithTimeout
	//				 https://pkg.go.dev/github.com/chromedp/chromedp#Run

	tabCtx, cancel1 := context.WithTimeout(browserCtx, time.Duration(chrome.Timeout) * time.Second)
	defer cancel1()

	// Run the initial browser
	if err := chromedp.Run(browserCtx); err != nil{
		return nil, err
	}

	var buf []byte

	// adding additional listeners to prevent browser crashes from locking the context (prevents hanging)
	chromedp.ListenTarget(browserCtx, func(ev interface{}) {
		if _, ok := ev.(*inspector.EventTargetCrashed); ok {
			go func() {
				cancel()
			}()
		}
	})
	
	chromedp.ListenTarget(tabCtx, func(ev interface{}) {
		if _, ok := ev.(*inspector.EventTargetCrashed); ok {
			go func() {
				cancel1()
			}()
		}
	})

	// squash JavaScript dialog boxes such as alert();
	chromedp.ListenTarget(tabCtx, func(ev interface{}) {
		if _, ok := ev.(*page.EventJavascriptDialogOpening); ok {
			go func() {
				if err := chromedp.Run(tabCtx,
					page.HandleJavaScriptDialog(true),
				); err != nil {
					cancel1()
				}
			}()
		}
	})

	// perform navigation on the tab context and attempt to take a clean screenshot
	if err := chromedp.Run(tabCtx, buildTasks(chrome,url,true,&buf)); err != nil {
		// if the context timeout exceeded (e.g. on a long page load) then just take the screenshot
		// this will take a screenshot of whatever loaded before failing

		// create a new tab context for this scenario, since our previous context expired
		// using a context timeout delay again to help prevent hanging scenarios
		tabCtxBackup, cancel2 := context.WithTimeout(browserCtx, time.Duration(chrome.Timeout) * time.Second)
		defer cancel2()

		// listen for crashes on this backup context as well
		chromedp.ListenTarget(tabCtxBackup, func(ev interface{}) {
			if _, ok := ev.(*inspector.EventTargetCrashed); ok {
				go func() {
					cancel2()
				}()
			}
		})

		// attempt to capture the screenshot of the tab
		if err2 := chromedp.Run(tabCtxBackup, buildTasks(chrome,url,false,&buf)); err2 != nil {
			return nil, err
		}
	}

	// return the screenshot buffer
	return buf, nil
}

// builds the chromedp tasks slice in a bit cleaner fashion with the various logic
func buildTasks(chrome *Chrome, url *url.URL, doNavigate bool, buf *[]byte) chromedp.Tasks {
	var actions chromedp.Tasks
	if doNavigate {
		actions = append(actions, chromedp.Navigate(url.String()))
		if chrome.Delay > 0 {
			actions = append(actions, chromedp.Sleep(time.Duration(chrome.Delay) * time.Second))
		}
		actions = append(actions, chromedp.Stop())
	}
	if chrome.FullPage{
		actions = append(actions,chromedp.FullScreenshot(buf,100))
	} else {
		actions = append(actions,chromedp.CaptureScreenshot(buf))
	}
	return actions
}