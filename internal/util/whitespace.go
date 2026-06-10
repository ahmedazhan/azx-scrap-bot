package util

import (
	"regexp"
	"strings"
)

var nonPrintable = regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`)

var wsCollapse = regexp.MustCompile(`\s+`)

func CollapseWS(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	return wsCollapse.ReplaceAllString(s, " ")
}

func TrimNonPrintable(s string) string {
	return nonPrintable.ReplaceAllString(s, "")
}
