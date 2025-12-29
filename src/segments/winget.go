package segments

import (
	"strings"

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
	var headerLine string
	separatorIndex := -1

	// First pass: find header line and separator
	for i, line := range lines {
		if line == "" {
			continue
		}

		if strings.Contains(line, "---") || strings.Contains(line, "─") {
			separatorIndex = i
			break
		}

		// Store the header line to determine column positions
		if strings.Contains(line, "Id") && strings.Contains(line, "Version") {
			headerLine = line
		}
	}

	// If no separator found, return empty
	if separatorIndex < 0 {
		return packages
	}

	// Determine column positions from header (calculated once)
	idIndex := strings.Index(headerLine, "Id")
	versionIndex := strings.Index(headerLine, "Version")
	availableIndex := strings.Index(headerLine, "Available")
	sourceIndex := strings.Index(headerLine, "Source")

	// If we can't find column positions, return empty
	if idIndex < 0 || versionIndex < 0 || availableIndex < 0 {
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

		// Skip lines that are too short
		if len(line) < availableIndex {
			continue
		}

		// Extract fields using column positions
		name := strings.TrimSpace(line[:idIndex])

		var id, current, available string
		if len(line) >= versionIndex {
			id = strings.TrimSpace(line[idIndex:versionIndex])
		}
		if len(line) >= availableIndex {
			current = strings.TrimSpace(line[versionIndex:availableIndex])
		}

		if sourceIndex > 0 && len(line) >= sourceIndex {
			available = strings.TrimSpace(line[availableIndex:sourceIndex])
		}

		if available == "" && len(line) > availableIndex {
			// No source column or line is shorter
			available = strings.TrimSpace(line[availableIndex:])
		}

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
