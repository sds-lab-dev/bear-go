package ui

import (
	"strings"
	"unicode"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	rw "github.com/mattn/go-runewidth"
	"github.com/rivo/uniseg"
)

// updateTextareaWithAutoResize wraps textarea.Update to prevent viewport
// scrolling issues caused by the textarea's internal repositionView().
//
// Problem: textarea.Update() calls repositionView() with the CURRENT height.
// When text wraps to a new visual line, the cursor moves beyond the visible
// area, and repositionView() scrolls the viewport down (increasing YOffset).
// Setting the correct height afterwards does not reset YOffset, so earlier
// lines become hidden.
//
// Solution: Before calling textarea.Update(), temporarily set the height to a
// value that guarantees the viewport's internal maxYOffset equals zero. This
// prevents repositionView() from scrolling at all. After Update returns, the
// correct height is set based on the actual visual line count.
//
// Why the chosen height works:
//
// The viewport's ScrollDown is clamped by maxYOffset:
//
//	maxYOffset = max(0, len(viewport.lines) - viewport.Height)
//
// To guarantee maxYOffset=0, we need viewport.Height >= len(viewport.lines).
// viewport.lines is populated by the previous View() cycle. The textarea's
// View() produces two groups of lines:
//
//  1. Content lines: the actual wrapped text (V lines, where V is the
//     textarea's internal visual line count from memoizedWrap)
//  2. Padding lines: always exactly m.height lines
//
// Each line ends with '\n', and strings.Split adds one extra element after
// the trailing newline. So: len(viewport.lines) = V + m.height + 1.
//
// We don't know V exactly (it's the textarea's internal wrap result, which
// may differ from our countVisualLines). However, V is bounded by the
// number of runes in the text plus 1 (worst case: every rune on its own
// line at width=1). Combined with the known m.height from the previous
// cycle, the upper bound is: len(runes) + 1 + previousHeight + 1.
//
// By setting the pre-expand height to this upper bound, we guarantee
// maxYOffset <= 0, making ScrollDown a no-op for all message types.
func updateTextareaWithAutoResize(ta textarea.Model, msg tea.Msg) (textarea.Model, tea.Cmd) {
	runeCount := len([]rune(ta.Value()))
	ta.SetHeight(runeCount + ta.Height() + 2)
	ta, cmd := ta.Update(msg)
	ta.SetHeight(countVisualLines(ta.Value(), ta.Width()))
	return ta, cmd
}

// countVisualLines calculates the total number of visual (display) lines
// for the given text at the specified wrap width. This accounts for
// soft-wrapping that occurs when a logical line exceeds the wrap width.
//
// The algorithm replicates the word-wrap logic used internally by the
// charmbracelet/bubbles textarea component, so the returned count matches
// the actual rendered line count in the textarea viewport.
func countVisualLines(text string, wrapWidth int) int {
	if wrapWidth <= 0 {
		return strings.Count(text, "\n") + 1
	}

	logicalLines := strings.Split(text, "\n")
	total := 0
	for _, line := range logicalLines {
		total += countWrappedLines([]rune(line), wrapWidth)
	}
	return total
}

// countWrappedLines counts how many visual rows a single logical line
// (without newlines) occupies when word-wrapped at the given width.
//
// This is a faithful port of the textarea package's internal wrap function,
// returning only the line count instead of building the actual wrapped content.
func countWrappedLines(runes []rune, width int) int {
	var (
		lines      = [][]rune{{}}
		word       []rune
		row        int
		spaceCount int
	)

	for _, r := range runes {
		if unicode.IsSpace(r) {
			spaceCount++
		} else {
			word = append(word, r)
		}

		if spaceCount > 0 {
			lineWidth := uniseg.StringWidth(string(lines[row]))
			wordWidth := uniseg.StringWidth(string(word))
			if lineWidth+wordWidth+spaceCount > width {
				row++
				lines = append(lines, []rune{})
				lines[row] = append(lines[row], word...)
				lines[row] = append(lines[row], repeatRune(' ', spaceCount)...)
				spaceCount = 0
				word = nil
			} else {
				lines[row] = append(lines[row], word...)
				lines[row] = append(lines[row], repeatRune(' ', spaceCount)...)
				spaceCount = 0
				word = nil
			}
		} else if len(word) > 0 {
			lastCharWidth := rw.RuneWidth(word[len(word)-1])
			if uniseg.StringWidth(string(word))+lastCharWidth > width {
				if len(lines[row]) > 0 {
					row++
					lines = append(lines, []rune{})
				}
				lines[row] = append(lines[row], word...)
				word = nil
			}
		}
	}

	lineWidth := uniseg.StringWidth(string(lines[row]))
	wordWidth := uniseg.StringWidth(string(word))
	if lineWidth+wordWidth+spaceCount >= width {
		row++
		lines = append(lines, []rune{})
		lines[row] = append(lines[row], word...)
	} else {
		lines[row] = append(lines[row], word...)
	}

	return len(lines)
}

func repeatRune(r rune, n int) []rune {
	result := make([]rune, n)
	for i := range result {
		result[i] = r
	}
	return result
}
