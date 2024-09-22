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

type submitSingleRequest struct {
	URL     string                `json:"url"`
	Options *submitRequestOptions `json:"options"`
}

// SubmitSingleHandler submits a URL to scan, returning the result.
//
//	@Summary		Submit a single URL for probing
//	@Description	Starts a new probing routine for a URL and options, returning the results when done.
//	@Tags			Results
//	@Accept			json
//	@Produce		json
//	@Param			query	body		submitSingleRequest	true	"The URL scanning request object"
//	@Success		200		{object}	models.Result		"The URL Result object"
//	@Router			/submit/single [post]
func (h *ApiHandler) SubmitSingleHandler(w http.ResponseWriter, r *http.Request) {
	var request submitSingleRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Error("failed to read json request", "err", err)
		http.Error(w, "Error reading JSON request", http.StatusInternalServerError)
		return
	}

	if request.URL == "" {
		http.Error(w, "No URL provided", http.StatusBadRequest)
		return
	}

	options := runner.NewDefaultOptions()
	options.Scan.ScreenshotToWriter = true
	options.Scan.ScreenshotSkipSave = true

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

	writer, err := writers.NewMemoryWriter(1)
	if err != nil {
		http.Error(w, "Error getting a memory writer", http.StatusInternalServerError)
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

	go func() {
		runner.Targets <- request.URL
		close(runner.Targets)
	}()

	runner.Run()
	runner.Close()

	jsonData, err := json.Marshal(writer.GetLatest())
	if err != nil {
		http.Error(w, "Error creating JSON response", http.StatusInternalServerError)
		return
	}

	w.Write(jsonData)
}
