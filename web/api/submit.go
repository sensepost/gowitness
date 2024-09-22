package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/sensepost/gowitness/pkg/log"
	"github.com/sensepost/gowitness/pkg/runner"
	driver "github.com/sensepost/gowitness/pkg/runner/drivers"
	"github.com/sensepost/gowitness/pkg/writers"
)

type submitRequest struct {
	URLs    []string              `json:"urls"`
	Options *submitRequestOptions `json:"options"`
}

type submitRequestOptions struct {
	X         int    `json:"window_x"`
	Y         int    `json:"window_y"`
	UserAgent string `json:"user_agent"`
	Timeout   int    `json:"timeout"`
	Delay     int    `json:"delay"`
	Format    string `json:"format"`
}

// SubmitHandler submits URL's for scans, writing them to the database.
//
//	@Summary		Submit URL's for scanning
//	@Description	Starts a new scanning routine for a list of URL's and options, writing results to the database.
//	@Tags			Results
//	@Accept			json
//	@Produce		json
//	@Param			query	body		submitRequest	true	"The URL scanning request object"
//	@Success		200		{string}	string			"Probing started"
//	@Router			/submit [post]
func (h *ApiHandler) SubmitHandler(w http.ResponseWriter, r *http.Request) {
	var request submitRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Error("failed to read json request", "err", err)
		http.Error(w, "Error reading JSON request", http.StatusInternalServerError)
		return
	}

	if len(request.URLs) == 0 {
		http.Error(w, "No URLs provided", http.StatusBadRequest)
		return
	}

	options := runner.NewDefaultOptions()
	options.Scan.ScreenshotPath = h.ScreenshotPath

	// Override default values with request options
	if request.Options != nil {
		if request.Options.X != 0 {
			options.Chrome.WindowX = request.Options.X
		}
		if request.Options.Y != 0 {
			options.Chrome.WindowY = request.Options.Y
		}
		if request.Options.UserAgent != "" {
			options.Chrome.UserAgent = request.Options.UserAgent
		}
		if request.Options.Timeout != 0 {
			options.Scan.Timeout = request.Options.Timeout
		}
		if request.Options.Delay != 0 {
			options.Scan.Delay = request.Options.Delay
		}
		if request.Options.Format != "" {
			options.Scan.ScreenshotFormat = request.Options.Format
		}
	}

	writer, err := writers.NewDbWriter(h.DbURI, false)
	if err != nil {
		http.Error(w, "Error connecting to DB for writer", http.StatusInternalServerError)
		return
	}

	logger := slog.New(log.Logger)

	driver, err := driver.NewChromedp(logger, *options)
	if err != nil {
		http.Error(w, "Error sarting driver", http.StatusInternalServerError)
		return
	}

	runner, err := runner.NewRunner(logger, driver, *options, []writers.Writer{writer})
	if err != nil {
		log.Error("error starting runner", "err", err)
		http.Error(w, "Error starting runner", http.StatusInternalServerError)
		return
	}

	// have everything we need! start ther runner goroutine
	go dispatchRunner(runner, request.URLs)

	response := `Probing started`
	jsonData, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error creating JSON response", http.StatusInternalServerError)
		return
	}

	w.Write(jsonData)
}

// dispatchRunner run's a runner in a separate goroutine
func dispatchRunner(runner *runner.Runner, targets []string) {
	// feed in targets
	go func() {
		for _, url := range targets {
			runner.Targets <- url
		}
		close(runner.Targets)
	}()

	runner.Run()
	runner.Close()
}
