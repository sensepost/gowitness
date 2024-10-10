package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
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
	Screenshot     string   `json:"screenshot"`
	MatchedFields  []string `json:"matched_fields"`
}

// searchOperators are the operators we support. everything else is
// "free text"
var searchOperators = []string{"title", "body", "tech", "header", "p"}

// SearchHandler handles search
//
//	@Summary		Search for results
//	@Description	Searches for results based on free form text, or operators.
//	@Tags			Results
//	@Accept			json
//	@Produce		json
//	@Param			query	body		searchRequest	true	"The search term to search for. Supports search operators: `title:`, `tech:`, `header:`, `body:`, `p:`"
//	@Success		200		{object}	searchResult
//	@Router			/search [post]
func (h *ApiHandler) SearchHandler(w http.ResponseWriter, r *http.Request) {
	var request searchRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Error("failed to read json request", "err", err)
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

		case "body":
			var bodyResults []models.Result
			if err := h.DB.Model(&models.Result{}).
				Where("LOWER(html) LIKE ?", lowerValue).Find(&bodyResults).Error; err != nil {
				log.Error("failed to get html results", "err", err)
				return
			}
			searchResults = appendResults(searchResults, resultIDs, bodyResults, key)

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
		case "p":
			var perceptionHashResults []models.Result
			if err := h.DB.Model(&models.Result{}).
				Where("perception_hash_group_id in (?)", h.DB.Model(&models.Result{}).
					Select("perception_hash_group_id").Distinct("perception_hash_group_id").
					Where(
						"perception_hash = ?",
						// p: was used as the operatator trigger, but we need it
						// back to resolve the group_id.
						fmt.Sprintf("p:%s", value),
					)).
				Find(&perceptionHashResults).Error; err != nil {

				log.Error("failed to get perception hash results", "err", err)
				return
			}

			searchResults = appendResults(searchResults, resultIDs, perceptionHashResults, key)
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

// parseSearchQuery parses a search query string into key-value pairs for known operators
// and captures any remaining free-form text.
func parseSearchQuery(query string) (map[string]string, string) {
	// Operators that we know of and that will be parsed

	result := make(map[string]string)

	var freeText string
	var currentKey string
	var currentValue []string

	parts := strings.Fields(query)

	for i := 0; i < len(parts); i++ {
		part := parts[i]

		// Check if the part contains an operator (e.g., title: or tech:)
		if index := strings.Index(part, ":"); index != -1 {
			operator := part[:index]
			if slices.Contains(searchOperators, operator) {
				// If we are processing an operator, finalize the previous key-value pair
				if currentKey != "" {
					result[currentKey] = strings.Join(currentValue, " ")
					currentValue = nil
				}
				// Set the current key to the new operator
				currentKey = operator

				// Handle the value right after the colon
				remainingPart := part[index+1:]
				// quoted value?
				if strings.HasPrefix(remainingPart, `"`) {
					// Quoted value (with spaces)
					remainingPart = strings.Trim(remainingPart, `"`)
					currentValue = append(currentValue, remainingPart)

					// Continue appending parts until the closing quote
					for i+1 < len(parts) && !strings.HasSuffix(parts[i+1], `"`) {
						i++
						currentValue = append(currentValue, parts[i])
					}
					if i+1 < len(parts) && strings.HasSuffix(parts[i+1], `"`) {
						i++
						closingPart := strings.Trim(parts[i], `"`)
						currentValue = append(currentValue, closingPart)
					}
				} else if remainingPart != "" {
					// Unquoted single word after colon
					currentValue = append(currentValue, remainingPart)
				} else if i+1 < len(parts) && !strings.HasPrefix(parts[i+1], `"`) {
					// Unquoted value in the next part
					i++
					currentValue = append(currentValue, parts[i])
				}
				continue
			}
		}

		// Add remaining parts as free text
		freeText += part + " "
	}

	// If we have an unprocessed key-value pair, store it
	if currentKey != "" {
		result[currentKey] = strings.Join(currentValue, " ")
	}

	// Trim any excess spaces from freeText
	freeText = strings.TrimSpace(freeText)

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
				Screenshot:     res.Screenshot,
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
