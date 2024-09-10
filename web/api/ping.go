package api

import (
	"encoding/json"
	"net/http"
)

// PingHandler handles ping requests
func (h *ApiHandler) PingHandler(w http.ResponseWriter, r *http.Request) {
	response := `pong`

	jsonData, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error creating JSON response", http.StatusInternalServerError)
		return
	}

	w.Write(jsonData)
}
