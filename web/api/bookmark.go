package api

import (
	"encoding/json"
	"net/http"

	"github.com/sensepost/gowitness/pkg/log"
)

type bookmarkRequest struct {
	ID int `json:"id"`
}

// BookmarkHandler marks a result as bookmarked.
//
//	@Summary		Bookmark results
//	@Description	Marks a given result as bookmarked, writing results to the database.
//	@Tags			Results
//	@Accept			json
//	@Produce		json
//	@Param			query	body	bookmarkRequest	true	"The bookmark request object"
//	@Success		200		{string}	string			"ok"
//	@Router			/results/bookmark [post]
func (h *ApiHandler) BookmarkHandler(w http.ResponseWriter, r *http.Request) {
	var request bookmarkRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Error("failed to read json request", "err", err)
		http.Error(w, "Error reading JSON request", http.StatusInternalServerError)
		return
	}

	log.Info("bookmarking id", "id", request.ID)

	if err := h.DB.Update("bookmark", true).Where("id = ?", request.ID).Error; err != nil {
		log.Error("failed to bookmark result", "err", err)
		return
	}

	response := `ok`
	jsonData, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error creating JSON response", http.StatusInternalServerError)
		return
	}

	w.Write(jsonData)
}
