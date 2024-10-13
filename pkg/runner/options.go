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
	// WSS is a websocket URL. Setting this will prevent gowitness
	// form launching Chrome, but rather use the remote instance.
	WSS string
	// Proxy server to use
	Proxy string
	// UserAgent is the user-agent string to set for Chrome
	UserAgent string
	// Headers to add to every request
	Headers []string
	// WindowSize, in pixels. Eg; X=1920,Y=1080
	WindowX int
	WindowY int
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
	Stdout    bool
	None      bool
}

// Scan is scanning related options
type Scan struct {
	// The scan driver to use. Can be one of [gorod, chromedp]
	Driver string
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
	// Don't write HTML response content
	SkipHTML bool
	// ScreenshotPath is the path where screenshot images will be stored.
	// An empty value means drivers will not write screenshots to disk. In
	// that case, you'd need to specify writer saves.
	ScreenshotPath string
	// ScreenshotFormat to save as
	ScreenshotFormat string
	// ScreenshotFullPage saves full, scrolled web pages
	ScreenshotFullPage bool
	// ScreenshotToWriter passes screenshots as a model property to writers
	ScreenshotToWriter bool
	// ScreenshotSkipSave skips saving screenshots to disk
	ScreenshotSkipSave bool
	// JavaScript to evaluate on every page
	JavaScript     string
	JavaScriptFile string
	// Save content stores content from network requests (warning) this
	// could make written artefacts huge
	SaveContent bool
}

// NewDefaultOptions returns Options with some default values
func NewDefaultOptions() *Options {
	return &Options{
		Chrome: Chrome{
			UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36",
			WindowX:   1920,
			WindowY:   1080,
		},
		Scan: Scan{
			Driver:           "chromedp",
			Threads:          6,
			Timeout:          60,
			UriFilter:        []string{"http", "https"},
			ScreenshotFormat: "jpeg",
		},
		Logging: Logging{
			Debug:         true,
			LogScanErrors: true,
		},
	}
}
