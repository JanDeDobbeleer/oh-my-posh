package segments

import (
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
)

const (
	fetchWindows options.Option = "fetch_windows"
)

// TmuxWindow holds information about a single tmux window.
type TmuxWindow struct {
	Index  string
	Name   string
	Active bool
}

// Tmux displays the current tmux session name and optionally the window list.
type Tmux struct {
	Base

	SessionName string
	Windows     []TmuxWindow
}

func (t *Tmux) Template() string {
	return " \ue7a2 {{ .SessionName }}{{ if .Windows }} | {{ range .Windows }}{{if .Active}}*{{end}}{{.Index}}:{{.Name}} {{end}}{{end}} "
}

func (t *Tmux) Enabled() bool {
	if !t.fetchSessionName() {
		return false
	}

	if t.options.Bool(fetchWindows, false) {
		t.Windows = t.fetchWindowList()
	}

	return true
}

func (t *Tmux) fetchSessionName() bool {
	// Try the tmux command for the exact session name (works when tmux is in PATH).
	if name, err := t.env.RunCommand("tmux", "display-message", "-p", "#S"); err == nil {
		t.SessionName = strings.TrimSpace(name)
		if t.SessionName != "" {
			return true
		}
	}

	// When running from the tmux status bar (#(...) format), the tmux binary may not
	// be in the minimal PATH used by /bin/sh. Use $TMUX to confirm we are inside tmux,
	// then fall back to the tmux format alias #S — tmux expands it to the actual session
	// name when processing the status bar format string.
	if t.env.Getenv("TMUX") == "" {
		return false
	}

	t.SessionName = "#S"
	return true
}

func (t *Tmux) fetchWindowList() []TmuxWindow {
	output, err := t.env.RunCommand("tmux", "list-windows", "-F", "#{window_index}\t#{window_name}\t#{window_active}")
	if err != nil {
		return nil
	}

	return t.parseWindows(output)
}

func (t *Tmux) parseWindows(output string) []TmuxWindow {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	windows := make([]TmuxWindow, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "\t")
		if len(parts) < 3 {
			continue
		}

		windows = append(windows, TmuxWindow{
			Index:  parts[0],
			Name:   parts[1],
			Active: parts[2] == "1",
		})
	}

	return windows
}
