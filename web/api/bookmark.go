package api

import (
	"encoding/json"
	"net/http"

	"github.com/sensepost/gowitness/pkg/log"
)

type bookmarkRequest struct {
	ID int `json:"id"`
}

// BookmarkHandler inverts the state of a bookmark
//
//	@Summary		Bookmark/Unbookmarks result
//	@Description	Inverts the bookmark status of a result, writing results to the database.
//	@Tags			Results
//	@Accept			json
//	@Produce		json
//	@Param			query	body	bookmarkRequest	true	"The bookmark request object"
//	@Success		200		{string}	string			"bookmarked"
//	@Router			/results/bookmark [post]
func (h *ApiHandler) BookmarkHandler(w http.ResponseWriter, r *http.Request) {
	var request bookmarkRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Error("failed to read json request", "err", err)
		http.Error(w, "Error reading JSON request", http.StatusInternalServerError)
		return
	}

	var bookmarked bool
	var err error
	bookmarkResult := h.DB.Select("bookmarked").Where("id = ?", request.ID).First(&bookmarked)
	if bookmarkResult.Error != nil {
		log.Error("failed to get bookmark status", "err", bookmarkResult.Error)
		http.Error(w, "Error getting result bookmark value", http.StatusInternalServerError)
		return
	}

	log.Info("inverting bookmark id", "id", request.ID)
	if err := h.DB.Update("bookmark", !bookmarked).Where("id = ?", request.ID).Error; err != nil {
		log.Error("failed to update result bookmark", "err", err)
		http.Error(w, "Error updating result bookmark value", http.StatusInternalServerError)
		return
	}

	var response string
	if bookmarked {
		response = `removed bookmark`
	} else {
		response = `bookmarked`
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error creating JSON response", http.StatusInternalServerError)
		return
	}

	w.Write(jsonData)
}
