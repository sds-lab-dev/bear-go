package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

func wrapWords(text string, maxWidth int) []string {
	if text == "" || maxWidth <= 0 {
		return nil
	}
	wrapped := lipgloss.Wrap(text, maxWidth, " ")
	return strings.Split(wrapped, "\n")
}

type wrappedStringBuilder struct {
	builder   strings.Builder
	lineWidth int
}

func newWrappedStringBuilder(maxWidth int) *wrappedStringBuilder {
	if maxWidth <= 0 {
		maxWidth = 80
	}
	return &wrappedStringBuilder{
		builder:   strings.Builder{},
		lineWidth: maxWidth,
	}
}

func (wsb *wrappedStringBuilder) WriteString(s string) (n int, err error) {
	return wsb.builder.WriteString(s)
}

func (wsb *wrappedStringBuilder) WriteByte(c byte) error {
	return wsb.builder.WriteByte(c)
}

func (wsb *wrappedStringBuilder) WriteRune(r rune) (int, error) {
	return wsb.builder.WriteRune(r)
}

func (wsb *wrappedStringBuilder) Write(p []byte) (int, error) {
	return wsb.builder.Write(p)
}

func (wsb *wrappedStringBuilder) String() string {
	return lipgloss.Wrap(wsb.builder.String(), wsb.lineWidth, " ")
}
