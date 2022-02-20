package chrome

import (
	"io"
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

func isTitleElement(n *html.Node) bool {
	return n.Type == html.ElementNode && n.Data == "title"
}

func traverse(n *html.Node) (string, bool) {

	if isTitleElement(n) {

		// handle empty <title> node
		if n.FirstChild == nil {
			return "(empty)", true
		}

		return n.FirstChild.Data, true
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result, ok := traverse(c)
		if ok {
			return strings.TrimSpace(result), ok
		}
	}

	return "", false
}

// GetHTMLTitle will parse the Title from an HTML document
// ref:
//	https://siongui.github.io/2016/05/10/go-get-html-title-via-net-html/
func GetHTMLTitle(r io.Reader) (string, bool) {
	doc, err := html.Parse(r)
	if err != nil {
		return "", false
	}

	return traverse(doc)
}

// GetTechnologies uses wapalyzer signatures to return an array
// of technologies that are in use by the remote site.
func GetTechnologies(headers http.Header, body []byte) []string {
	var technologies []string

	fingerprints := wappalyzerClient.Fingerprint(headers, body)

	for match := range fingerprints {
		technologies = append(technologies, match)
	}

	return technologies
}
