package chrome

import (
	"context"
	"crypto/tls"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"
	"net"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/sensepost/gowitness/storage"
	"gorm.io/gorm"
	"golang.org/x/net/proxy"
)

// Chrome contains information about a Google Chrome
// instance, with methods to run on it.
type Chrome struct {
	ResolutionX int
	ResolutionY int
	UserAgent   string
	Timeout     int64
	FullPage    bool
	ChromePath  string
	ProxyServer string
}

// NewChrome returns a new initialised Chrome struct
func NewChrome() *Chrome {
	return &Chrome{}
}

// Preflight will preflight a url
func (chrome *Chrome) Preflight(url *url.URL) (resp *http.Response, title string, err error) {
	// proxy setup for preflight
	baseDialer := &net.Dialer {
		Timeout:	time.Duration(chrome.Timeout)*time.Second,
		KeepAlive:	time.Duration(chrome.Timeout)*time.Second,
	}
	dialContext := (baseDialer).DialContext
	if chrome.ProxyServer != "" {
		// url, could use error handling for bad urls
		pUrl, _ := url.Parse(chrome.ProxyServer)
		// proxy.FromURL supports both HTTP and socks5 protocols
		dialSocksProxy, _ := proxy.FromURL(pUrl, baseDialer)
		if contextDialer, ok := dialSocksProxy.(proxy.ContextDialer); ok {
			dialContext = contextDialer.DialContext
		}
	}

	// purposefully ignore bad certs
	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
			DisableKeepAlives: true,
			Proxy:		   http.ProxyFromEnvironment,
			DialContext:	   dialContext,
		},
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

	return
}

// StorePreflight will store preflight info to a DB
func (chrome *Chrome) StorePreflight(url *url.URL, db *gorm.DB, resp *http.Response, title string, filename string) (uint, error) {

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

	if chrome.ProxyServer != "" {
		options = append(options, chromedp.ProxyServer(chrome.ProxyServer))
	}

	actx, acancel := chromedp.NewExecAllocator(context.Background(), options...)
	ctx, cancel := chromedp.NewContext(actx)
	defer acancel()
	defer cancel()

	var buf []byte

	if chrome.FullPage {
		// straight from: https://github.com/chromedp/examples/blob/255873ca0d76b00e0af8a951a689df3eb4f224c3/screenshot/main.go#L54
		if err := chromedp.Run(ctx, chromedp.Tasks{
			chromedp.Navigate(url.String()),
			chromedp.ActionFunc(func(ctx context.Context) error {
				// get layout metrics
				_, _, contentSize, err := page.GetLayoutMetrics().Do(ctx)
				if err != nil {
					return err
				}

				width, height := int64(math.Ceil(contentSize.Width)),
					int64(math.Ceil(contentSize.Height))

				// force viewport emulation
				err = emulation.SetDeviceMetricsOverride(width, height, 1, false).
					WithScreenOrientation(&emulation.ScreenOrientation{
						Type:  emulation.OrientationTypePortraitPrimary,
						Angle: 0,
					}).Do(ctx)
				if err != nil {
					return err
				}

				// capture screenshot
				buf, err = page.CaptureScreenshot().
					WithQuality(100).
					WithClip(&page.Viewport{
						X:      contentSize.X,
						Y:      contentSize.Y,
						Width:  contentSize.Width,
						Height: contentSize.Height,
						Scale:  2,
					}).Do(ctx)
				if err != nil {
					return err
				}
				return nil
			}),
		}); err != nil {
			return nil, err
		}

	} else {
		// normal viewport screenshot
		if err := chromedp.Run(ctx, chromedp.Tasks{
			chromedp.Navigate(url.String()),
			chromedp.CaptureScreenshot(&buf),
		}); err != nil {
			return nil, err
		}
	}

	return buf, nil
}
