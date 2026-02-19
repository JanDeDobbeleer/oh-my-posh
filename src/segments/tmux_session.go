package segments

import "strings"

// TmuxSession displays the current tmux session name.
type TmuxSession struct {
	Base

	SessionName string
}

func (t *TmuxSession) Template() string {
	return " \ue7a2 {{ .SessionName }} "
}

func (t *TmuxSession) Enabled() bool {
	// Prefer tmux's own display-message so we get the exact session name.
	if name, err := t.env.RunCommand("tmux", "display-message", "-p", "#S"); err == nil {
		t.SessionName = strings.TrimSpace(name)
		return t.SessionName != ""
	}

	// Fallback: parse $TMUX which has the format "/tmp/tmux-NNN/socket,PID,sessionIndex".
	// The session index is not the name, but it is better than nothing.
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
