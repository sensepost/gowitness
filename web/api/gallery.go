package api

import (
	"encoding/json"
	"net/http"
	"strconv"

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
	ID           uint     `json:"id"`
	URL          string   `json:"url"`
	ResponseCode int      `json:"response_code"`
	Title        string   `json:"title"`
	Filename     string   `json:"file_name"`
	Failed       bool     `json:"failed"`
	Technologies []string `json:"technologies"`
}

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

	// query the db
	var queryResults []*models.Result
	if err := h.DB.Model(&models.Result{}).Limit(results.Limit).
		Offset(offset).Preload("Technologies").Find(&queryResults).Error; err != nil {

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
			URL:          result.URL,
			ResponseCode: result.ResponseCode,
			Title:        result.Title,
			Filename:     result.Filename,
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
