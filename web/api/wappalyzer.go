package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// probably not the smartest idea, but he
const iconBase = `https://raw.githubusercontent.com/enthec/webappanalyzer/main/src/images/icons/`

func (h *ApiHandler) WappalyzerHandler(w http.ResponseWriter, r *http.Request) {
	response := make(map[string]string)

	for name, finger := range h.Wappalyzer.GetFingerprints().Apps {
		if finger.Icon == "" {
			continue
		}

		response[name] = fmt.Sprintf(`%s%s`, iconBase, finger.Icon)
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(jsonData)
}
