package chrome

import (
	"context"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/pkg/errors"

	gover "github.com/mcuadros/go-version"
	log "github.com/sirupsen/logrus"
)

// Chrome contains information about a Google Chrome
// instance, with methods to run on it.
type Chrome struct {
	Resolution string

	path           string
	screenshotPath string
}

// Setup configures a Chrome struct with the path
// specified to what is available on this system.
func (chrome *Chrome) Setup() {

	// Ensure we have a resolution set
	if chrome.Resolution == "" {
		chrome.Resolution = "1440,900"
	}

	chrome.chromeLocator()
	chrome.checkVersion()
}

// ChromeLocator looks for an installation of Google Chrome
// and returns the path to where the installation was found
func (chrome *Chrome) chromeLocator() {

	// Possible paths for Google Chrome or chromium to be at.
	paths := []string{
		"/usr/bin/chromium",
		"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		"/Applications/Google Chrome Canary.app/Contents/MacOS/Google Chrome Canary",
		"/Applications/Chromium.app/Contents/MacOS/Chromium",
		"C:/Program Files (x86)/Google/Chrome/Application/chrome.exe",
	}

	for _, path := range paths {

		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}

		log.WithField("chrome-path", path).Debug("Google Chrome path")
		chrome.path = path
	}
}

// testVersion gets the version of Google Chrome that we have
func (chrome *Chrome) checkVersion() {

	out, err := exec.Command(chrome.path, "-version").Output()
	if err != nil {
		log.WithField("err", err).Fatal("An error occured while trying to get the Chrome version")
	}

	// Convert the output to a simple string
	version := string(out)

	re := regexp.MustCompile(`\d+(\.\d+)+`)
	match := re.FindStringSubmatch(version)
	if len(match) <= 0 {
		log.Warn("Unable to determine Chrome version.")
		return
	}

	// grab the first match in the version extraction
	version = match[0]

	if gover.Compare(version, "60", "<") {
		log.WithField("chromeversion", version).Fatal("Chrome version is older than v60")
	}

	log.WithField("version", version).Debug("Chrome version")
}

// ScreenshotPath sets the path for screenshots
func (chrome *Chrome) ScreenshotPath(p string) error {

	p, err := filepath.Abs(p)
	if err != nil {
		return err
	}

	if _, err := os.Stat(p); os.IsNotExist(err) {
		return errors.New("Destination path does not exist")
	}

	log.WithField("screenshot-path", p).Debug("Screenshot path")
	chrome.screenshotPath = p

	return nil
}

// ScreenshotURL takes a screenshot of a URL
func (chrome *Chrome) ScreenshotURL(targetURL *url.URL, destination string) {

	screenshotLocation := filepath.Join(chrome.screenshotPath, destination)
	log.WithFields(log.Fields{"url": targetURL, "full-destination": screenshotLocation}).
		Debug("Full path to screenshot save")

	// Start with the basic headless arguments
	var chromeArguments = []string{
		"--headless", "--disable-gpu", "--hide-scrollbars",
		"--disable-crash-reporter",
		"--user-agent='Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 " +
			"(KHTML, like Gecko) Chrome/60.0.3112.50 Safari/537.36'",
		"--window-size=" + chrome.Resolution, "--screenshot=" + screenshotLocation,
	}

	// When we are running as root, chromiun will flag the 'cant
	// run as root' thing. Handle that case.
	if os.Geteuid() == 0 {

		log.WithField("euid", os.Geteuid()).Debug("Running as root, adding --no-sandbox")
		chromeArguments = append(chromeArguments, "--no-sandbox")
	}

	// Check if we need to add a proxy hack for Chrome headless to
	// stfu about certificates :>
	if targetURL.Scheme == "https" {

		// Chrome headless... you suck. Proxy to the target
		// so that we can ignore SSL certificate issues.
		// proxy := shittyProxy{targetURL: targetURL}
		proxy := forwardingProxy{targetURL: targetURL}

		// Give the shitty proxy a few moments to start up.
		time.Sleep(500 * time.Millisecond)

		// Start the proxy and grab the listening port we should tell
		// Chrome to connect to.
		if err := proxy.start(); err != nil {

			log.WithField("error", err).Warning("Failed to start proxy for HTTPS request")
			return
		}

		// Update the URL scheme back to http, the proxy will handle the SSL
		proxyURL, _ := url.Parse("http://localhost:" + strconv.Itoa(proxy.port) + "/")
		proxyURL.Path = targetURL.Path

		// I am not 100% sure if this does anything, but lets add --allow-insecure-localhost
		// anyways.
		chromeArguments = append(chromeArguments, "--allow-insecure-localhost")

		// set the URL to call to the proxy we are starting up
		chromeArguments = append(chromeArguments, proxyURL.String())

		// when we are done, stop the hack :|
		defer proxy.stop()

	} else {

		// Finally add the url to screenshot
		chromeArguments = append(chromeArguments, targetURL.String())
	}

	log.WithFields(log.Fields{"arguments": chromeArguments}).Debug("Google Chrome arguments")

	// get a context to run the command in
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	// Prepare the command to run...
	cmd := exec.CommandContext(ctx, chrome.path, chromeArguments...)

	log.WithFields(log.Fields{"url": targetURL, "destination": destination}).Info("Taking screenshot")

	// ... and run it!
	startTime := time.Now()
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	// Wait for the screenshot to finish and handle the error that may occur.
	if err := cmd.Wait(); err != nil {

		// If if this error was as a result of a timeout
		if ctx.Err() == context.DeadlineExceeded {
			log.WithFields(log.Fields{"url": targetURL, "destination": destination, "err": err}).
				Error("Timeout reached while waiting for screenshot to finish")
			return
		}

		log.WithFields(log.Fields{"url": targetURL, "destination": destination, "err": err}).
			Error("Screenshot failed")

		return
	}

	log.WithFields(log.Fields{
		"url": targetURL, "destination": destination, "duration": time.Since(startTime),
	}).Info("Screenshot taken")
}
