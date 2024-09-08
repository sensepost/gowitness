package runner

// Options are global gowitness options
type Options struct {
	// Logging is logging options
	Logging Logging
	// Chrome is Chrome related options
	Chrome Chrome
	// Writer is output options
	Writer Writer
	// Scan is typically Scan options
	Scan Scan
}

// Logging is log related options
type Logging struct {
	// Debug display debug level logging
	Debug bool
	// LogScanErrors log errors related to scanning
	LogScanErrors bool
	// Silence all logging
	Silence bool
}

// Chrome is Google Chrome related options
type Chrome struct {
	// Path to the Chrome binary. An empty value implies that
	// go-rod will auto download a platform appropriate binary
	// to use.
	Path string
	// Proxy server to use
	Proxy string
	// UserAgent is the user-agent string to set for Chrome
	UserAgent string
	// Headers to add to every request
	Headers []string
	// WindowSize, in pixels. Eg; 1920,1080
	WindowSize string
}

// Writer options
type Writer struct {
	Db        bool
	DbURI     string
	DbDebug   bool // enables verbose database logs
	Csv       bool
	CsvFile   string
	Jsonl     bool
	JsonlFile string
}

// Scan is scanning related options
type Scan struct {
	// Threads (not really) are the number of goroutines to use.
	// More soecifically, its the go-rod page pool well use.
	Threads int
	// Timeout is the maximum time to wait for a page load before timing out.
	Timeout int
	// Number of seconds of delay between navigation and screenshotting
	Delay int
	// UriFilter are URI's that are okay to process. This should normally
	// be http and https
	UriFilter []string
	// ScreenshotPath is the path where screenshot images will be stored
	ScreenshotPath string
	// ScreenshotFormat to save as
	ScreenshotFormat string
	// ScreenshotFullPage saves full, scrolled web pages
	ScreenshotFullPage bool
	// JavaScript to evaluate on every page
	JavaScript     string
	JavaScriptFile string
}