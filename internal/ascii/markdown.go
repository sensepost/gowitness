package ascii

import (
	"strings"

	"github.com/charmbracelet/glamour"
)

var renderer *glamour.TermRenderer

// Markdown renders markdown
func Markdown(s string) string {
	r, err := renderer.Render(strings.TrimSpace(s))
	if err != nil {
		panic(err)
	}

	return r
}

func init() {
	var err error
	renderer, err = glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithPreservedNewLines(),
	)
	if err != nil {
		panic(err)
	}
}
