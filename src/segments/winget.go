package segments

import (
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
)

type WinGet struct {
	Base

	UpdateCount int
	Updates     []WinGetPackage
}

type WinGetPackage struct {
	Name      string
	ID        string
	Current   string
	Available string
}

const (
	WINGETCACHEKEY = "winget_updates"
)

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

	duration := w.options.String(options.CacheDuration, string(cache.ONEDAY))

	updates, ok := cache.Get[[]WinGetPackage](cache.Device, WINGETCACHEKEY)
	if ok {
		w.Updates = updates
		w.UpdateCount = len(updates)
		return w.UpdateCount > 0
	}

	output, err := w.env.RunCommand("winget", "upgrade")
	if err != nil {
		return false
	}

	w.Updates = w.parseWinGetOutput(output)
	w.UpdateCount = len(w.Updates)

	cache.Set(cache.Device, WINGETCACHEKEY, w.Updates, cache.Duration(duration))

	return w.UpdateCount > 0
}

func (w *WinGet) parseWinGetOutput(output string) []WinGetPackage {
	var packages []WinGetPackage

	lines := strings.Split(output, "\n")
	var headerLine string
	separatorFound := false

	for _, line := range lines {
		if line == "" {
			continue
		}

		// Find the header line and separator
		if !separatorFound {
			if strings.Contains(line, "---") || strings.Contains(line, "─") {
				separatorFound = true
				continue
			}
			// Store the header line to determine column positions
			if strings.Contains(line, "Id") && strings.Contains(line, "Version") {
				headerLine = line
			}
			continue
		}

		// Determine column positions from header
		idIndex := strings.Index(headerLine, "Id")
		versionIndex := strings.Index(headerLine, "Version")
		availableIndex := strings.Index(headerLine, "Available")
		sourceIndex := strings.Index(headerLine, "Source")

		// If we can't find column positions, skip
		if idIndex < 0 || versionIndex < 0 || availableIndex < 0 {
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
		} else if len(line) > availableIndex {
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
