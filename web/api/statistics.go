package api

import (
	"encoding/json"
	"net/http"

	"github.com/sensepost/gowitness/pkg/log"
	"github.com/sensepost/gowitness/pkg/models"
)

type statisticsResponse struct {
	DbSize        int64                     `json:"dbsize"`
	Results       int64                     `json:"results"`
	Headers       int64                     `json:"headers"`
	NetworkLogs   int64                     `json:"networklogs"`
	ConsoleLogs   int64                     `json:"consolelogs"`
	ResponseCodes []*statisticsResponseCode `json:"response_code_stats"`
}

type statisticsResponseCode struct {
	Code  int   `json:"code"`
	Count int64 `json:"count"`
}

// StatisticsHandler returns database statistics
//
//	@Summary		Database statistics
//	@Description	Get database statistics.
//	@Tags			Results
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	statisticsResponse
//	@Router			/statistics [get]
func (h *ApiHandler) StatisticsHandler(w http.ResponseWriter, r *http.Request) {
	response := &statisticsResponse{}

	if err := h.DB.Raw("SELECT page_count * page_size as size FROM pragma_page_count(), pragma_page_size()").
		Take(&response.DbSize).Error; err != nil {

		log.Error("an error occured getting database size", "err", err)
		return
	}

	if err := h.DB.Model(&models.Result{}).Count(&response.Results).Error; err != nil {
		log.Error("an error occured counting results", "err", err)
		return
	}

	if err := h.DB.Model(&models.Header{}).Count(&response.Headers).Error; err != nil {
		log.Error("an error occured counting headers", "err", err)
		return
	}

	if err := h.DB.Model(&models.NetworkLog{}).Count(&response.NetworkLogs).Error; err != nil {
		log.Error("an error occured counting network logs", "err", err)
		return
	}

	if err := h.DB.Model(&models.ConsoleLog{}).Count(&response.ConsoleLogs).Error; err != nil {
		log.Error("an error occured counting console logs", "err", err)
		return
	}

	var counts []*statisticsResponseCode
	if err := h.DB.Model(&models.Result{}).
		Select("response_code as code, count(*) as count").
		Group("response_code").Scan(&counts).Error; err != nil {
		log.Error("failed counting response codes", "err", err)
		return
	}

	response.ResponseCodes = counts

	jsonData, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(jsonData)
}
