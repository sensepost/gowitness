package utils

import (
	"crypto/tls"
	"net/url"
	"strings"
	"time"

	chrm "github.com/leonjza/gowitness/chrome"
	"github.com/parnurzeal/gorequest"
	log "github.com/sirupsen/logrus"
)

const (
	// HTTP is the prefix for http:// urls
	HTTP string = "http://"
	// HTTPS is the prefox for https:// urls
	HTTPS string = "https://"
)

// ProcessURL processes a URL
func ProcessURL(url *url.URL, chrome *chrm.Chrome, timeout int) {

	log.WithField("url", url).Debug("Processing URL")

	request := gorequest.New().Timeout(time.Duration(timeout)*time.Second).
		TLSClientConfig(&tls.Config{InsecureSkipVerify: true}).
		Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.50 Safari/537.36")

	resp, _, errs := request.Get(url.String()).End()
	if errs != nil {
		log.WithFields(log.Fields{"url": url, "error": errs}).Debug("Failed to query url")

		return
	}

	log.WithFields(log.Fields{"url": url, "status": resp.Status}).Info("Response code")

	finalURL := resp.Request.URL
	log.WithFields(log.Fields{"url": url, "final-url": finalURL}).Info("Final URL after redirects")

	for k, v := range resp.Header {
		log.WithFields(log.Fields{"url": url, k: strings.Join(v, ", ")}).Info("Response header")
	}

	if resp.TLS != nil {
		for _, c := range resp.TLS.PeerCertificates {
			log.WithFields(log.Fields{"url": url, "common_name": c.Subject.CommonName}).Info("Certificate chain common name")
			log.WithFields(log.Fields{"url": url, "signature-alg": c.SignatureAlgorithm}).Info("Signature algorithm")
			log.WithFields(log.Fields{"url": url, "pubkey-alg": c.PublicKeyAlgorithm}).Info("Public key algorithm")
			log.WithFields(log.Fields{"url": url, "issuer": c.Issuer.CommonName}).Info("Issuer")

			for _, d := range c.DNSNames {

				log.WithFields(log.Fields{"url": url, "dns-names": d}).Info("DNS Name")
			}
		}
		log.WithFields(log.Fields{"url": url, "cipher-suite": resp.TLS.CipherSuite}).Info("Cipher suite in use")
	}

	dst := SafeFileName(url.String()) + ".png"
	log.WithFields(log.Fields{"url": url, "destination": dst}).Debug("Generated filename for screenshot")

	chrome.ScreenshotURL(finalURL, dst)
}
