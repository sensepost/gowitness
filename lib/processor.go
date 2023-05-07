package lib

import (
	"bytes"
	"image/png"
	"net/url"
	"os"

	"github.com/corona10/goimagehash"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sensepost/gowitness/chrome"
	"github.com/sensepost/gowitness/storage"
	"gorm.io/gorm"
)

// Processor is a URL processing helper
type Processor struct {
	Logger *zerolog.Logger

	Db                 *gorm.DB
	Chrome             *chrome.Chrome
	URL                *url.URL
	ScreenshotPath     string
	ScreenshotFileName string

	// file name & file path
	fn string
	fp string

	// preflight response
	preflightResult *chrome.PreflightResult
	// screenshot
	screenshotResult *chrome.ScreenshotResult

	// persistence id
	urlid uint
}

// Gowitness processes a URL by:
//   - preflighting
//   - storing
//   - screenshotting
//   - calculating a perception hash
//   - writing a screenshot to disk
func (p *Processor) Gowitness() (err error) {

	p.init()

	if err = p.preflight(); err != nil {
		log.Error().Err(err).Msg("preflight request failed")
		return
	}

	// check if the preflight returned a code to process.
	// an empty slice implies no filtering
	if (len(p.Chrome.ScreenshotCodes) > 0) &&
		!SliceContainsInt(p.Chrome.ScreenshotCodes, p.preflightResult.HTTPResponse.StatusCode) {

		log.Warn().Int("response-code", p.preflightResult.HTTPResponse.StatusCode).
			Msg("response code not in allowed screenshot http response codes. skipping.")

		return
	}

	if err = p.takeScreenshot(); err != nil {
		log.Error().Err(err).Msg("failed to take screenshot")
		return
	}

	if err = p.persistRequest(); err != nil {
		log.Error().Err(err).Msg("failed to store request information")
		return
	}

	if err = p.storePerceptionHash(); err != nil {
		log.Error().Err(err).Msg("failed to calculate and save a perception hash")
		return
	}

	if err = p.writeScreenshot(); err != nil {
		log.Error().Err(err).Msg("failed to save screenshot buffer")
		return
	}

	return
}

// adds the file extension to a given path if the path does not end with the given extension
func addExtensionIfNeeded(filepath, extension string) string {
	ext := filepath[len(filepath)-len(extension):]
	if ext != extension {
		return filepath[:] + extension
	}
	return filepath
}

// init initialises the Processor
func (p *Processor) init() {
	if p.ScreenshotFileName != "" {
		p.fn = p.ScreenshotFileName
	} else {
		p.fn = SafeFileName(p.URL.String())
	}

	// limit filename length
	p.fn = TruncateString(p.fn, 30)

	// set the extention depending on the screenshot format
	if p.Chrome.AsPDF {
		p.fn = addExtensionIfNeeded(p.fn, ".pdf")
	} else {
		p.fn = addExtensionIfNeeded(p.fn, ".png")
	}

	p.fp = ScreenshotPath(p.fn, p.URL, p.ScreenshotPath)
}

// preflight invokes the Chrome preflight helper
func (p *Processor) preflight() (err error) {
	p.Logger.Debug().Str("url", p.URL.String()).Msg("preflighting")

	p.preflightResult, err = p.Chrome.Preflight(p.URL)
	if err != nil {
		return
	}

	var l *zerolog.Event
	if p.preflightResult.HTTPResponse.StatusCode == 200 {
		l = p.Logger.Info()
	} else {
		l = p.Logger.Warn()
	}
	l.Str("url", p.URL.String()).Int("statuscode", p.preflightResult.HTTPResponse.StatusCode).
		Str("title", p.preflightResult.HTTPTitle).Msg("preflight result")

	return
}

// persistRequest dispatches the StorePreflight function
func (p *Processor) persistRequest() (err error) {

	if p.Db == nil {
		return
	}

	p.Logger.Debug().Str("url", p.URL.String()).Msg("storing request data")
	if p.urlid, err = p.Chrome.StoreRequest(p.Db, p.preflightResult, p.screenshotResult, p.fn); err != nil {
		return
	}

	return
}

// takeScreenshot dispatches the takeScreenshot function
func (p *Processor) takeScreenshot() (err error) {
	p.Logger.Debug().Str("url", p.URL.String()).Msg("screenshotting")

	p.screenshotResult, err = p.Chrome.Screenshot(p.URL)
	if err != nil {
		return
	}

	return
}

// storePerceptionHash calculates and stores a perception hash
func (p *Processor) storePerceptionHash() (err error) {

	if p.Db == nil {
		return
	}

	// ignore pdf's
	if p.Chrome.AsPDF {
		return
	}

	p.Logger.Debug().Str("url", p.URL.String()).Msg("calculating perception hash")
	img, err := png.Decode(bytes.NewReader(p.screenshotResult.Screenshot))
	if err != nil {
		return
	}

	comp, err := goimagehash.PerceptionHash(img)
	if err != nil {
		return
	}

	var dburl storage.URL
	p.Db.First(&dburl, p.urlid)
	dburl.PerceptionHash = comp.ToString()
	p.Db.Save(&dburl)

	return
}

// writeScreenshot writes the screenshot buffer to disk
func (p *Processor) writeScreenshot() (err error) {

	p.Logger.Debug().Str("url", p.URL.String()).Str("path", p.fp).Msg("saving screenshot buffer")
	if err = os.WriteFile(p.fp, p.screenshotResult.Screenshot, 0644); err != nil {
		return
	}

	return
}
