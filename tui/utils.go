package tui

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

func deleteLastRune(s string) string {
	if s == "" {
		return s
	}
	_, size := utf8.DecodeLastRuneInString(s)
	if size <= 0 || size > len(s) {
		return s
	}
	return s[:len(s)-size]
}

func deleteLastWord(s string) string {
	s = strings.TrimRightFunc(s, unicode.IsSpace)
	for s != "" {
		r, size := utf8.DecodeLastRuneInString(s)
		if size <= 0 || size > len(s) {
			break
		}
		if unicode.IsSpace(r) {
			break
		}
		s = s[:len(s)-size]
	}
	return s
}

