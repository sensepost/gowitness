package options

// Options are global gowitness options
type Options struct {
	// Logging is logging options
	Logging Logging
	// Output is output options
	Output Output
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

// Output options
type Output struct {
	Db        bool
	DbFile    string
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
	// UriFilter are URI's that are okay to process. This should normally
	// be http and https
	UriFilter []string
	// ScreenshotPath is the path where screenshot images will be stored
	ScreenshotPath string
	// UserAgent is the user-agent string to set for Chrome
	UserAgent string
}
