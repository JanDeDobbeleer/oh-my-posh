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
	// Prefer tmux's own display-message so we get the exact session name.
	if name, err := t.env.RunCommand("tmux", "display-message", "-p", "#S"); err == nil {
		t.SessionName = strings.TrimSpace(name)
		return t.SessionName != ""
	}

	// Fallback: parse $TMUX which has the format "/tmp/tmux-NNN/socket,PID,windowIndex".
	// The window index is not the session name, but it is a last-resort identifier
	// when tmux display-message is unavailable.
	tmuxEnv := t.env.Getenv("TMUX")
	if tmuxEnv == "" {
		return false
	}

	parts := strings.Split(tmuxEnv, ",")
	if len(parts) >= 3 {
		t.SessionName = parts[2]
		return t.SessionName != ""
	}

	return false
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
