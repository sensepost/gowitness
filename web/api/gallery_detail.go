package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sensepost/gowitness/pkg/log"
	"github.com/sensepost/gowitness/pkg/models"
	"gorm.io/gorm/clause"
)

func (h *ApiHandler) GalleryDetailHandler(w http.ResponseWriter, r *http.Request) {
	var response = &models.Result{}

	v := h.DB.Model(&models.Result{}).
		Preload(clause.Associations).
		Preload("TLS.SanList").
		First(&response, chi.URLParam(r, "id"))
	if v.Error != nil {
		log.Error("could not get detail for id", "err", v.Error)
		return
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(jsonData)
}
