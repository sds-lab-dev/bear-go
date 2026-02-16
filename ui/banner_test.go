package ui

import (
	"strings"
	"testing"
)

func stripANSI(s string) string {
	var result strings.Builder
	inEscape := false
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}
		result.WriteRune(r)
	}
	return result.String()
}

func TestRenderBanner_ContainsBearArt(t *testing.T) {
	banner := RenderBanner(120)
	plain := stripANSI(banner)

	artFragments := []string{
		"() _ _ ()",
		"/@@\\",
		"________",
	}

	for _, fragment := range artFragments {
		if !strings.Contains(plain, fragment) {
			t.Errorf("banner should contain bear art fragment %q", fragment)
		}
	}
}

func TestRenderBanner_ContainsSlogan(t *testing.T) {
	banner := RenderBanner(120)
	plain := stripANSI(banner)

	if !strings.Contains(plain, "Bear: The AI developer that saves your time.") {
		t.Error("banner should contain slogan text")
	}
}

func TestRenderBanner_ContainsDescription(t *testing.T) {
	banner := RenderBanner(120)
	plain := stripANSI(banner)

	if !strings.Contains(plain, "heavy lifting") {
		t.Error("banner should contain description text fragment")
	}
}

func TestRenderBanner_ContainsSeparator(t *testing.T) {
	banner := RenderBanner(120)

	if !strings.Contains(banner, "─") {
		t.Error("banner should contain horizontal separator")
	}
}

func TestRenderBanner_NarrowTerminal(t *testing.T) {
	wideRightCol := buildRightColumn(120 - bearArtWidth)
	narrowRightCol := buildRightColumn(60 - bearArtWidth)

	if len(narrowRightCol) <= len(wideRightCol) {
		t.Errorf(
			"narrow terminal should produce more right-column lines: wide=%d, narrow=%d",
			len(wideRightCol), len(narrowRightCol),
		)
	}
}
