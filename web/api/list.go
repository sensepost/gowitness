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

func (h *ApiHandler) ListHandler(w http.ResponseWriter, r *http.Request) {
	var results = []*listResponse{}

	v := h.DB.Model(&models.Result{}).Find(&results)
	if v.Error != nil {
		log.Error("could not get list", "err", v.Error)
		return
	}

	// v = h.DB.Model(&models.Result{}).Count(&results.TotalCount)
	// if v.Error != nil {
	// 	log.Error("could not count total results", "err", v.Error)
	// 	return
	// }

	jsonData, err := json.Marshal(results)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(jsonData)
}
