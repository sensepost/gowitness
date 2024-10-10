package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sensepost/gowitness/pkg/log"
	"github.com/sensepost/gowitness/pkg/models"
)

type galleryResponse struct {
	Results    []*galleryContent `json:"results"`
	Page       int               `json:"page"`
	Limit      int               `json:"limit"`
	TotalCount int64             `json:"total_count"`
}

type galleryContent struct {
	ID           uint      `json:"id"`
	ProbedAt     time.Time `json:"probed_at"`
	URL          string    `json:"url"`
	ResponseCode int       `json:"response_code"`
	Title        string    `json:"title"`
	Filename     string    `json:"file_name"`
	Screenshot   string    `json:"screenshot"`
	Failed       bool      `json:"failed"`
	Technologies []string  `json:"technologies"`
}

// GalleryHandler gets a paginated gallery
//
//	@Summary		Gallery
//	@Description	Get a paginated list of results.
//	@Tags			Results
//	@Accept			json
//	@Produce		json
//	@Param			page			query		int		false	"The page to load."
//	@Param			limit			query		int		false	"Number of results per page."
//	@Param			technologies	query		string	false	"A comma seperated list of technologies to filter by."
//	@Param			status			query		string	false	"A comma seperated list of HTTP status codes to filter by."
//	@Param			perception		query		boolean	false	"Order the results by perception hash."
//	@Param			failed			query		boolean	false	"Include failed screenshots in the results."
//	@Success		200				{object}	galleryResponse
//	@Router			/results/gallery [get]
func (h *ApiHandler) GalleryHandler(w http.ResponseWriter, r *http.Request) {
	var results = &galleryResponse{
		Page:  1,
		Limit: 24,
	}

	// pagination
	urlPage := r.URL.Query().Get("page")
	urlLimit := r.URL.Query().Get("limit")
	if p, err := strconv.Atoi(urlPage); err == nil && p > 0 {
		results.Page = p
	}
	if l, err := strconv.Atoi(urlLimit); err == nil && l > 0 {
		results.Limit = l
	}
	offset := (results.Page - 1) * results.Limit

	// perception sorting
	var perceptionSort bool
	perceptionSortValue := r.URL.Query().Get("perception")
	perceptionSort, err := strconv.ParseBool(perceptionSortValue)
	if err != nil {
		perceptionSort = false
	}

	// status code filtering
	var statusCodes []int
	statusFilterValue := r.URL.Query().Get("status")
	if statusFilterValue != "" {
		for _, statusCodeString := range strings.Split(statusFilterValue, ",") {
			statusCode, err := strconv.Atoi(statusCodeString)
			if err != nil {
				continue
			}

			statusCodes = append(statusCodes, statusCode)
		}
	}

	// technology filtering
	var technologies []string
	technologyFilterValue := r.URL.Query().Get("technologies")
	if technologyFilterValue != "" {
		technologies = append(technologies, strings.Split(technologyFilterValue, ",")...)
	}

	// failed result filtering
	var showFailed bool
	showFailed, err = strconv.ParseBool(r.URL.Query().Get("failed"))
	if err != nil {
		showFailed = true
	}

	// query the db
	var queryResults []*models.Result
	query := h.DB.Model(&models.Result{}).Limit(results.Limit).
		Offset(offset).Preload("Technologies")

	if perceptionSort {
		query.Order("perception_hash_group_id DESC")
	}

	if len(statusCodes) > 0 {
		query.Where("response_code IN ?", statusCodes)
	}

	if len(technologies) > 0 {
		query.Where("id in (?)", h.DB.Model(&models.Technology{}).
			Select("result_id").Distinct("result_id").
			Where("value IN (?)", technologies))
	}

	if !showFailed {
		query.Where("failed = ?", showFailed)
	}

	// run the query
	if err := query.Find(&queryResults).Error; err != nil {
		log.Error("could not get gallery", "err", err)
		return
	}

	// extract Technologies for each result
	for _, result := range queryResults {
		var technologies []string
		for _, tech := range result.Technologies {
			technologies = append(technologies, tech.Value)
		}

		// Append the processed data to the response
		results.Results = append(results.Results, &galleryContent{
			ID:           result.ID,
			ProbedAt:     result.ProbedAt,
			URL:          result.URL,
			ResponseCode: result.ResponseCode,
			Title:        result.Title,
			Filename:     result.Filename,
			Screenshot:   result.Screenshot,
			Failed:       result.Failed,
			Technologies: technologies,
		})
	}

	if err := h.DB.Model(&models.Result{}).Count(&results.TotalCount).Error; err != nil {
		log.Error("could not count total results", "err", err)
		return
	}

	jsonData, err := json.Marshal(results)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(jsonData)
}
