package api

import (
	wappalyzer "github.com/projectdiscovery/wappalyzergo"
	"github.com/sensepost/gowitness/pkg/database"
	"gorm.io/gorm"
)

// ApiHandler is an API handler
type ApiHandler struct {
	DbURI          string
	ScreenshotPath string
	DB             *gorm.DB
	Wappalyzer     *wappalyzer.Wappalyze
}

// NewApiHandler returns a new ApiHandler
func NewApiHandler(uri string, screenshotPath string) (*ApiHandler, error) {

	// get a db handle
	conn, err := database.Connection(uri, false, false)
	if err != nil {
		return nil, err
	}

	wap, _ := wappalyzer.New()

	return &ApiHandler{
		DbURI:          uri,
		ScreenshotPath: screenshotPath,
		DB:             conn,
		Wappalyzer:     wap,
	}, nil
}
