package web

import (
	"net/http"
)

//go:generate go run ../generate_assets.go

// AssetsLocation is the path to web assets
var AssetsLocation http.FileSystem = http.Dir("assets")
