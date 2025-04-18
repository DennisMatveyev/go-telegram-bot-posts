package app

import (
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

var multiNewlines = regexp.MustCompile(`\n{2,}`)

func cleanText(raw string) string {
	p := bluemonday.StrictPolicy()
	noHTML := p.Sanitize(raw)
	collapsed := multiNewlines.ReplaceAllString(noHTML, "\n")

	return strings.TrimSpace(collapsed)
}
