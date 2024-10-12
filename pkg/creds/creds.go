package creds

import "strings"

type Credential struct {
	Name        string   `json:"name"`
	Patterns    []string `json:"patterns"`
	Credentials []string `json:"credentials"`
	References  []string `json:"references"`
}

// Find potential credentials matching an HTML input
func Find(html string) []*Credential {
	var results = []*Credential{}

	for _, cred := range Credentials {
		for _, pat := range cred.Patterns {
			if strings.Contains(strings.ToLower(html), strings.ToLower(pat)) {
				results = append(results, cred)
				break
			}
		}
	}

	return results
}
