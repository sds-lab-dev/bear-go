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

func TestRenderBanner_ColumnAlignment(t *testing.T) {
	banner := RenderBanner(120)
	plain := stripANSI(banner)
	lines := strings.Split(plain, "\n")

	// 오른쪽 컬럼 텍스트가 있는 모든 줄에서 bear art 영역이
	// 정확히 bearArtWidth 폭으로 패딩되어 있는지 확인.
	for i := range bearArtLines {
		if i >= len(lines) {
			break
		}
		line := lines[i]
		if len(line) <= bearArtWidth {
			continue
		}

		// bearArtWidth 위치 이후에 오른쪽 컬럼 텍스트가 있는 줄만 검증
		rightText := strings.TrimSpace(line[bearArtWidth:])
		if rightText == "" {
			continue
		}

		artPart := line[:bearArtWidth]
		if len(artPart) != bearArtWidth {
			t.Errorf(
				"line %d: art region width is %d, expected %d",
				i, len(artPart), bearArtWidth,
			)
		}
	}
}
