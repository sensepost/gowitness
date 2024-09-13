package api

import (
	"encoding/json"
	"net/http"
)

// PingHandler handles ping requests
//
//	@Summary		Ping the server
//	@Description	Returns a simple "pong" response to test server availability.
//	@Tags			Health
//	@Accept			json
//	@Produce		json
//	@Success		200	{string}	string	"pong"
//	@Router			/api/ping [get]
func (h *ApiHandler) PingHandler(w http.ResponseWriter, r *http.Request) {
	response := `pong`

	jsonData, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error creating JSON response", http.StatusInternalServerError)
		return
	}

	w.Write(jsonData)
}
