package utils

import (
	"regexp"
	"strings"
)

// SafeFileName return a safe string that can be used in file names
func SafeFileName(str string) string {

	name := strings.ToLower(str)
	name = strings.Trim(name, " ")

	separators, err := regexp.Compile(`[ &_=+:]`)
	if err == nil {
		name = separators.ReplaceAllString(name, "-")
	}

	legal, err := regexp.Compile(`[^[:alnum:]-.]`)
	if err == nil {
		name = legal.ReplaceAllString(name, "")
	}

	for strings.Contains(name, "--") {
		name = strings.Replace(name, "--", "-", -1)
	}

	return name
}
