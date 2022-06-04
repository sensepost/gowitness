package chrome

import (
	"bytes"
	"net/http"
	"strings"

	wappalyzer "github.com/projectdiscovery/wappalyzergo"
	"golang.org/x/net/html"
)

type Wappalyzer struct {
	client *wappalyzer.Wappalyze
	err    error
}

// NewWappalyzer returns a new Wappalyzer instance.
// If an error occured, .error would contain an error()
func NewWappalyzer() *Wappalyzer {

	c, err := wappalyzer.New()

	// this is a hard error, and probably means the bundled
	// json from which matches are loaded has an error.
	// return a null Wappalyzer with an error.
	if err != nil {
		return &Wappalyzer{
			err: err,
		}
	}

	return &Wappalyzer{
		client: c,
	}
}

// Technologies uses wappalyzergo to determine known technologies from headers
// or an HTTP body
func (w *Wappalyzer) Technologies(headers http.Header, body []byte) (tech []string) {

	fingerprints := w.client.Fingerprint(headers, body)

	for match := range fingerprints {
		tech = append(tech, match)
	}

	return
}

// HTMLTitle returns the title parsed from an HTML <title> tag.
func (w *Wappalyzer) HTMLTitle(b []byte) string {

	r := bytes.NewReader(b)

	doc, err := html.Parse(r)
	if err != nil {
		return ""
	}

	title, _ := w.traverse(doc)

	return title
}

func (w *Wappalyzer) traverse(n *html.Node) (string, bool) {

	if w.isTitleElement(n) {

		// handle empty <title> node
		if n.FirstChild == nil {
			return "(empty)", true
		}

		return n.FirstChild.Data, true
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result, ok := w.traverse(c)
		if ok {
			return strings.TrimSpace(result), ok
		}
	}

	return "", false
}

func (w *Wappalyzer) isTitleElement(n *html.Node) bool {
	return n.Type == html.ElementNode && n.Data == "title"
}
