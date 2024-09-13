package api

import (
	"encoding/json"
	"net/http"

	"github.com/sensepost/gowitness/pkg/log"
	"github.com/sensepost/gowitness/pkg/models"
)

type listResponse struct {
	ID uint `json:"id" gorm:"primarykey"`

	URL            string `json:"url"`
	FinalURL       string `json:"final_url"`
	ResponseCode   int    `json:"response_code"`
	ResponseReason string `json:"response_reason"`
	Protocol       string `json:"protocol"`
	ContentLength  int64  `json:"content_length"`
	Title          string `json:"title"`

	// Failed flag set if the result should be considered failed
	Failed       bool   `json:"failed"`
	FailedReason string `json:"failed_reason"`
}

// ListHandler returns a simple list of results
//
//	@Summary		Results list
//	@Description	Get a simple list of all results.
//	@Tags			Results
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	listResponse
//	@Router			/results/list [get]
func (h *ApiHandler) ListHandler(w http.ResponseWriter, r *http.Request) {
	var results = []*listResponse{}

	if err := h.DB.Model(&models.Result{}).Find(&results).Error; err != nil {
		log.Error("could not get list", "err", err)
		return
	}

	jsonData, err := json.Marshal(results)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(jsonData)
}
