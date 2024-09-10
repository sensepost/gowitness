package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/sensepost/gowitness/pkg/log"
	"github.com/sensepost/gowitness/pkg/models"
)

type searchRequest struct {
	Query string `json:"query"`
}

type searchResult struct {
	ID uint `json:"id" gorm:"primarykey"`

	URL            string   `json:"url"`
	FinalURL       string   `json:"final_url"`
	ResponseCode   int      `json:"response_code"`
	ResponseReason string   `json:"response_reason"`
	Protocol       string   `json:"protocol"`
	ContentLength  int64    `json:"content_length"`
	Title          string   `json:"title"`
	Failed         bool     `json:"failed"`
	FailedReason   string   `json:"failed_reason"`
	Filename       string   `json:"file_name"`
	MatchedFields  []string `json:"matched_fields"`
}

// SearchHandler handles search
func (h *ApiHandler) SearchHandler(w http.ResponseWriter, r *http.Request) {
	var request searchRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Error reading JSON request", http.StatusInternalServerError)
		return
	}

	parsed, freeText := parseSearchQuery(request.Query)
	var searchResults []searchResult
	resultIDs := make(map[uint]bool)

	// iterate over parsed search operators
	for key, value := range parsed {
		lowerValue := fmt.Sprintf("%%%s%%", value)

		switch key {
		case "title":
			var titleResults []models.Result
			if err := h.DB.Model(&models.Result{}).Where("LOWER(title) LIKE ?", lowerValue).
				Find(&titleResults).Error; err != nil {

				log.Error("failed to get title results", "err", err)
				return
			}

			searchResults = appendResults(searchResults, resultIDs, titleResults, key)
		case "tech":
			var techResults []models.Result
			if err := h.DB.Model(&models.Result{}).
				Where("id in (?)", h.DB.Model(&models.Technology{}).
					Select("result_id").Distinct("result_id").
					Where("value LIKE ?", lowerValue)).
				Find(&techResults).Error; err != nil {

				log.Error("failed to get tech results", "err", err)
				return
			}

			searchResults = appendResults(searchResults, resultIDs, techResults, key)
		case "header":
			var headerResults []models.Result
			if err := h.DB.Model(&models.Result{}).
				Where("id in (?)", h.DB.Model(&models.Header{}).
					Select("result_id").Distinct("result_id").
					Where("value LIKE ?", lowerValue)).
				Find(&headerResults).Error; err != nil {

				log.Error("failed to get tech results", "err", err)
				return
			}

			searchResults = appendResults(searchResults, resultIDs, headerResults, key)
		}
	}

	// process any freetext if there is
	if freeText != "" {
		lowerFreeText := fmt.Sprintf("%%%s%%", freeText)
		var freeTextResults []models.Result

		if err := h.DB.Model(&models.Result{}).
			Where("LOWER(url) LIKE ?", lowerFreeText).
			Or("LOWER(final_url) LIKE ?", lowerFreeText).
			Or("LOWER(title) LIKE ?", lowerFreeText).
			Find(&freeTextResults).Error; err != nil {

			log.Error("failed to get freetext results", "err", err)
			return
		}

		searchResults = appendResults(searchResults, resultIDs, freeTextResults, "text")
	}

	jsonData, err := json.Marshal(searchResults)
	if err != nil {
		http.Error(w, "Error creating JSON response", http.StatusInternalServerError)
		return
	}

	w.Write(jsonData)
}

// parseSearchQuery will extract operator and value pairs
// Parse the search query into a map of operators and their values, and capture free-form text
func parseSearchQuery(query string) (map[string]string, string) {
	result := make(map[string]string)
	var freeText string

	// Regular expression to extract key-value pairs (e.g., title: "foo bar", header: bar)
	// Matches: key: "value with spaces" or key: valueWithoutSpaces
	re := regexp.MustCompile(`(\w+):\s*(?:"([^"]+)"|([^\s]+))`)

	// Extract all the key-value pairs from the query
	matches := re.FindAllStringSubmatch(query, -1)

	processedLength := 0

	for _, match := range matches {
		key := match[1]
		value := match[2] // match[2] is the quoted value (if any)
		if value == "" {
			value = match[3] // match[3] is the unquoted value
		}
		result[key] = strings.Trim(value, `"'`)
		processedLength += len(match[0]) + 1 // Keep track of processed characters
	}

	// Ensure that processedLength does not exceed the query length
	if processedLength > len(query) {
		processedLength = len(query)
	}

	// If there's any remaining free-form text after the operators
	remainingText := query[processedLength:]
	if strings.TrimSpace(remainingText) != "" {
		freeText = strings.TrimSpace(remainingText)
	}

	return result, freeText
}

// appendResults adds results to searchResults, ensuring unique results are added,
// and also tracks which field caused the match
func appendResults(searchResults []searchResult, resultIDs map[uint]bool, newResults []models.Result, matchedField string) []searchResult {
	for _, res := range newResults {
		if resultIDs[res.ID] {
			for i := range searchResults {
				if searchResults[i].ID == res.ID {
					searchResults[i].MatchedFields = appendUnique(searchResults[i].MatchedFields, matchedField)
					break
				}
			}
		} else {
			searchResults = append(searchResults, searchResult{
				ID:             res.ID,
				URL:            res.URL,
				FinalURL:       res.FinalURL,
				ResponseCode:   res.ResponseCode,
				ResponseReason: res.ResponseReason,
				Protocol:       res.Protocol,
				ContentLength:  res.ContentLength,
				Title:          res.Title,
				Failed:         res.Failed,
				FailedReason:   res.FailedReason,
				Filename:       res.Filename,
				MatchedFields:  []string{matchedField},
			})

			// Mark the result ID as added
			resultIDs[res.ID] = true
		}
	}
	return searchResults
}

// appendUnique ensures no duplicates in the list of matched fields
func appendUnique(existingFields []string, newField string) []string {
	for _, field := range existingFields {
		if field == newField {
			return existingFields
		}
	}
	return append(existingFields, newField)
}
