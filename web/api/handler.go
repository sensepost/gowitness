package api

import (
	wappalyzer "github.com/projectdiscovery/wappalyzergo"
	"gorm.io/gorm"
)

// ApiHandler is an API handler
type ApiHandler struct {
	DB         *gorm.DB
	Wappalyzer *wappalyzer.Wappalyze
}

// NewApiHandler returns a new ApiHandler
func NewApiHandler(db *gorm.DB) *ApiHandler {
	wap, _ := wappalyzer.New()
	return &ApiHandler{
		DB:         db,
		Wappalyzer: wap,
	}
}
