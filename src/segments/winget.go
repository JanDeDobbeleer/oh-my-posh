package segments

import (
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
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

	duration := w.props.GetString(properties.CacheDuration, string(cache.ONEDAY))

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
	separatorFound := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		// Find the separator line (contains dashes)
		if !separatorFound {
			if strings.Contains(line, "---") || strings.Contains(line, "─") {
				separatorFound = true
			}
			continue
		}

		// Parse package lines after separator
		fields := strings.Fields(line)

		// Need at least 4 fields: Name, ID, Current Version, Available Version
		if len(fields) < 4 {
			continue
		}

		// Skip any footer lines
		if strings.Contains(line, "upgrade") && strings.Contains(line, "available") {
			continue
		}

		pkg := WinGetPackage{
			Name:      fields[0],
			ID:        fields[1],
			Current:   fields[2],
			Available: fields[3],
		}

		packages = append(packages, pkg)
	}

	return packages
}
