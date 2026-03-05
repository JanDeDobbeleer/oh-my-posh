package segments

import (
	"strings"
	"unicode"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

type WinGet struct {
	Base

	Updates     []WinGetPackage
	UpdateCount int
}

type WinGetPackage struct {
	Name      string
	ID        string
	Current   string
	Available string
}

func (w *WinGet) Template() string {
	return " \uf409 {{ .UpdateCount }} "
}

func (w *WinGet) Enabled() bool {
	if w.env.GOOS() != runtime.WINDOWS {
		return false
	}

	if !w.env.HasCommand("winget") {
		return false
	}

	output, err := w.env.RunCommand("winget", "upgrade")
	if err != nil {
		return false
	}

	w.Updates = w.parseWinGetOutput(output)
	w.UpdateCount = len(w.Updates)

	return w.UpdateCount > 0
}

func (w *WinGet) parseWinGetOutput(output string) []WinGetPackage {
	var packages []WinGetPackage

	lines := strings.Split(output, "\n")
	separatorIndex := -1

	// Find the separator line
	for i, line := range lines {
		if line == "" {
			continue
		}

		if strings.Contains(line, "---") || strings.Contains(line, "─") {
			separatorIndex = i
			break
		}
	}

	// If no separator found, return empty
	if separatorIndex < 0 {
		return packages
	}

	// The header line is right before the separator
	if separatorIndex == 0 {
		return packages
	}

	headerLine := lines[separatorIndex-1]

	// Find column positions by detecting transitions from non-letter to letter
	columnIndices := findColumnIndices(headerLine)

	// We need at least 4 columns: Name, Id, Version, Available
	if len(columnIndices) < 4 {
		return packages
	}

	// Second pass: process data lines after separator
	for i := separatorIndex + 1; i < len(lines); i++ {
		line := lines[i]

		if line == "" {
			continue
		}

		// Skip footer lines
		if strings.Contains(line, "upgrade") && strings.Contains(line, "available") {
			continue
		}

		// Skip lines that are too short to contain all fields
		if len(line) < columnIndices[3] {
			continue
		}

		// Extract fields using column positions
		name := extractField(line, columnIndices[0], columnIndices[1])
		id := extractField(line, columnIndices[1], columnIndices[2])
		current := extractField(line, columnIndices[2], columnIndices[3])
		available := extractField(line, columnIndices[3], columnIndices[4])

		// Skip if essential fields are empty
		if name == "" || id == "" {
			continue
		}

		pkg := WinGetPackage{
			Name:      name,
			ID:        id,
			Current:   current,
			Available: available,
		}

		packages = append(packages, pkg)
	}

	return packages
}

func findColumnIndices(headerLine string) []int {
	var allIndices []int
	prevWasLetter := false

	for i, char := range headerLine {
		isLetter := unicode.IsLetter(char)

		if isLetter && !prevWasLetter {
			// Found transition from non-letter to letter - this is a column start
			allIndices = append(allIndices, i)
		}

		prevWasLetter = isLetter
	}

	// Get the last 5 column positions (Name, Id, Version, Available, Source)
	// This handles cases where there's garbage before the actual header
	lastFive := allIndices[len(allIndices)-5:]

	// Make all positions relative to the first column position
	firstColPos := lastFive[0]
	relativeIndices := make([]int, len(lastFive))
	for i, pos := range lastFive {
		relativeIndices[i] = pos - firstColPos
	}

	return relativeIndices
}

func extractField(line string, startIndex, endIndex int) string {
	if startIndex >= len(line) {
		return ""
	}
	if endIndex > len(line) {
		endIndex = len(line)
	}
	return strings.TrimSpace(line[startIndex:endIndex])
}
