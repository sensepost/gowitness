package lib

import (
	"bytes"
	"image/png"
	"io/ioutil"
	"net/http"
	"net/url"

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
	response     *http.Response
	title        string
	technologies []string
	// persistence id
	urlid uint
	// screenshot
	screenshot *[]byte
}

// Gowitness processes a URL by:
//	- preflighting
//	- storing
//	- screenshotting
//	- calculating a perception hash
//	- writing a screenshot to disk
func (p *Processor) Gowitness() (err error) {

	p.init()

	if err = p.preflight(); err != nil {
		log.Error().Err(err).Msg("preflight request failed")
		return
	}

	if err = p.persistPreflight(); err != nil {
		log.Error().Err(err).Msg("failed to store preflight information")
		return
	}

	if err = p.takeScreenshot(); err != nil {
		log.Error().Err(err).Msg("failed to take screenshot")
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

// init initialises the Processor
func (p *Processor) init() {
	if p.ScreenshotFileName != "" {
		p.fn = p.ScreenshotFileName
	} else {
		p.fn = SafeFileName(p.URL.String())
	}
	p.fp = ScreenshotPath(p.fn, p.URL, p.ScreenshotPath)
}

// preflight invokes the Chrome preflight helper
func (p *Processor) preflight() (err error) {
	p.Logger.Debug().Str("url", p.URL.String()).Msg("preflighting")
	p.response, p.title, p.technologies, err = p.Chrome.Preflight(p.URL)
	if err != nil {
		return
	}

	var l *zerolog.Event
	if p.response.StatusCode == 200 {
		l = p.Logger.Info()
	} else {
		l = p.Logger.Warn()
	}
	l.Str("url", p.URL.String()).Int("statuscode", p.response.StatusCode).
		Str("title", p.title).Msg("preflight result")

	return
}

// persistPreflight dispatches the StorePreflight function
func (p *Processor) persistPreflight() (err error) {

	if p.Db == nil {
		return
	}

	p.Logger.Debug().Str("url", p.URL.String()).Msg("storing preflight data")
	if p.urlid, err = p.Chrome.StorePreflight(p.URL, p.Db, p.response, p.title, p.technologies, p.fn); err != nil {
		return
	}

	return
}

// takeScreenshot dispatches the takeScreenshot function
func (p *Processor) takeScreenshot() (err error) {
	p.Logger.Debug().Str("url", p.URL.String()).Msg("screenshotting")
	buf, err := p.Chrome.Screenshot(p.URL)
	if err != nil {
		return
	}

	p.screenshot = &buf

	return
}

// storePerceptionHash calculates and stores a perception hash
func (p *Processor) storePerceptionHash() (err error) {

	if p.Db == nil {
		return
	}

	p.Logger.Debug().Str("url", p.URL.String()).Msg("calculating perception hash")
	img, err := png.Decode(bytes.NewReader(*p.screenshot))
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
	p.Logger.Debug().Str("url", p.URL.String()).Str("path", p.fn).Msg("saving screenshot buffer")
	if err = ioutil.WriteFile(p.fp, *p.screenshot, 0644); err != nil {
		return
	}

	return
}
