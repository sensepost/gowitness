package api

import (
	"encoding/json"
	"net/http"

	"github.com/sensepost/gowitness/pkg/log"
	"github.com/sensepost/gowitness/pkg/models"
)

type statisticsResponse struct {
	DbSize      int64 `json:"dbsize"`
	Results     int64 `json:"results"`
	Headers     int64 `json:"headers"`
	NetworkLogs int64 `json:"networklogs"`
	ConsoleLogs int64 `json:"consolelogs"`
}

func (h *ApiHandler) StatisticsHandler(w http.ResponseWriter, r *http.Request) {
	response := &statisticsResponse{}

	v := h.DB.Raw("SELECT page_count * page_size as size FROM pragma_page_count(), pragma_page_size()").Take(&response.DbSize)
	if v.Error != nil {
		log.Error("an error occured getting database size", "err", v.Error)
		return
	}

	v = h.DB.Model(&models.Result{}).Count(&response.Results)
	if v.Error != nil {
		log.Error("an error occured counting results", "err", v.Error)
		return
	}

	v = h.DB.Model(&models.Header{}).Count(&response.Headers)
	if v.Error != nil {
		log.Error("an error occured counting headers", "err", v.Error)
		return
	}

	v = h.DB.Model(&models.NetworkLog{}).Count(&response.NetworkLogs)
	if v.Error != nil {
		log.Error("an error occured counting network logs", "err", v.Error)
		return
	}

	v = h.DB.Model(&models.ConsoleLog{}).Count(&response.ConsoleLogs)
	if v.Error != nil {
		log.Error("an error occured counting console logs", "err", v.Error)
		return
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(jsonData)
}
