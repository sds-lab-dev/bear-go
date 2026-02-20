package ui

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

var bearArtLines = []string{
	"",
	`       () _ _ ()`,
	`      / __  __ \`,
	`/@@\ /  o    o  \ /@@\`,
	`\ @ \|     ^    |/ @ /`,
	` \   \    ___   /   /`,
	`  \   \________/   /`,
}

var bearArtWidth = calcBearArtWidth()

const (
	rightColumnStartOffset = 1
	bearArtGap             = 3
	sloganText             = "Bear: The AI developer that saves your time."
	descriptionText        = "Bear does the heavy lifting for you; you just collect your paycheck and don't worry about a thing."
)

func calcBearArtWidth() int {
	maxWidth := 0
	for _, line := range bearArtLines {
		if w := utf8.RuneCountInString(line); w > maxWidth {
			maxWidth = w
		}
	}
	return maxWidth
}

type styleName int

const (
	styleNone styleName = iota
	styleSlogan
	styleDescription
)

type rightColumnLine struct {
	text  string
	style styleName
}

func buildRightColumn(maxWidth int) []rightColumnLine {
	sloganLines := wrapWords(sloganText, maxWidth)
	descLines := wrapWords(descriptionText, maxWidth)

	result := make([]rightColumnLine, 0, len(sloganLines)+1+len(descLines))

	for _, line := range sloganLines {
		result = append(result, rightColumnLine{text: line, style: styleSlogan})
	}

	if len(sloganLines) > 0 && len(descLines) > 0 {
		result = append(result, rightColumnLine{text: "", style: styleNone})
	}

	for _, line := range descLines {
		result = append(result, rightColumnLine{text: line, style: styleDescription})
	}

	return result
}

func RenderBanner(terminalWidth int) string {
	rightColWidth := terminalWidth - bearArtWidth - bearArtGap
	rightColLines := buildRightColumn(rightColWidth)

	var builder strings.Builder

	for i, artLine := range bearArtLines {
		padded := fmt.Sprintf("%-*s", bearArtWidth+bearArtGap, artLine)
		builder.WriteString(BearArtStyle.Render(padded))

		rightIdx := i - rightColumnStartOffset
		if rightIdx >= 0 && rightIdx < len(rightColLines) {
			entry := rightColLines[rightIdx]
			styled := applyRightColumnStyle(entry)
			builder.WriteString(styled)
		}

		builder.WriteByte('\n')
	}

	separator := strings.Repeat("─", terminalWidth)
	builder.WriteString(SeparatorStyle.Render(separator))
	builder.WriteByte('\n')

	return builder.String()
}

func applyRightColumnStyle(entry rightColumnLine) string {
	switch entry.style {
	case styleSlogan:
		return SloganStyle.Render(entry.text)
	case styleDescription:
		return DescriptionStyle.Render(entry.text)
	default:
		return entry.text
	}
}
