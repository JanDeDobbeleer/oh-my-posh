package segments

import (
	"fmt"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/color"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"
)

const (
	tmuxActiveForeground   options.Option = "active_foreground"
	tmuxActiveBackground   options.Option = "active_background"
	tmuxInactiveForeground options.Option = "inactive_foreground"
	tmuxInactiveBackground options.Option = "inactive_background"
	tmuxPowerlineSymbol    options.Option = "powerline_symbol"

	defaultTmuxPowerlineSymbol = "\ue0b0"
)

type tmuxWindow struct {
	Index  string
	Name   string
	Active bool
}

// TmuxWindowList renders all tmux windows as a single powerline-connected string.
// It builds ANSI sequences directly rather than calling terminal.Write() to avoid
// racing with other segment goroutines that share the global terminal builder.
type TmuxWindowList struct {
	Base

	RenderedList string
}

func (t *TmuxWindowList) Template() string {
	return "{{ .RenderedList }}"
}

func (t *TmuxWindowList) Enabled() bool {
	output, err := t.env.RunCommand("tmux", "list-windows", "-F", "#{window_index}\t#{window_name}\t#{window_active}")
	if err != nil {
		return false
	}

	windows := t.parseWindows(output)
	if len(windows) == 0 {
		return false
	}

	activeFG := t.options.Color(tmuxActiveForeground, color.Ansi("#282828"))
	activeBG := t.options.Color(tmuxActiveBackground, color.Ansi("#83a598"))
	inactiveFG := t.options.Color(tmuxInactiveForeground, color.Ansi("#928374"))
	inactiveBG := t.options.Color(tmuxInactiveBackground, color.Ansi("#3c3836"))
	symbol := t.options.String(tmuxPowerlineSymbol, defaultTmuxPowerlineSymbol)

	if terminal.Plain {
		t.RenderedList = t.renderPlain(windows)
		return true
	}

	t.RenderedList = t.renderWindows(windows, activeFG, activeBG, inactiveFG, inactiveBG, symbol)
	return true
}

func (t *TmuxWindowList) parseWindows(output string) []tmuxWindow {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	windows := make([]tmuxWindow, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "\t")
		if len(parts) < 3 {
			continue
		}

		windows = append(windows, tmuxWindow{
			Index:  parts[0],
			Name:   parts[1],
			Active: parts[2] == "1",
		})
	}

	return windows
}

// renderWindows builds the full powerline-connected ANSI string for all windows.
// This method accesses terminal.Colors (read-only) and terminal.Plain but does NOT
// call terminal.Write(), keeping it safe to call from concurrent goroutines.
func (t *TmuxWindowList) renderWindows(windows []tmuxWindow, activeFG, activeBG, inactiveFG, inactiveBG color.Ansi, symbol string) string {
	if terminal.Colors == nil {
		return t.renderPlain(windows)
	}

	toFGSeq := func(c color.Ansi) string {
		resolved := terminal.Colors.ToAnsi(c, false)
		if resolved.IsEmpty() {
			return ""
		}
		return fmt.Sprintf("\x1b[%sm", resolved)
	}

	toBGSeq := func(c color.Ansi) string {
		resolved := terminal.Colors.ToAnsi(c, true)
		if resolved.IsEmpty() {
			return ""
		}
		return fmt.Sprintf("\x1b[%sm", resolved)
	}

	type winColors struct{ fg, bg color.Ansi }
	getColors := func(active bool) winColors {
		if active {
			return winColors{activeFG, activeBG}
		}
		return winColors{inactiveFG, inactiveBG}
	}

	var sb strings.Builder
	var prevBG color.Ansi // zero value = empty → no previous window

	for _, win := range windows {
		c := getColors(win.Active)

		if !prevBG.IsEmpty() {
			// Powerline separator: use prevBG as fg color, curBG as bg color.
			sb.WriteString(toFGSeq(prevBG))
			sb.WriteString(toBGSeq(c.bg))
			sb.WriteString(symbol)
		}

		sb.WriteString(toBGSeq(c.bg))
		sb.WriteString(toFGSeq(c.fg))
		sb.WriteString(fmt.Sprintf(" %s:%s ", win.Index, win.Name))

		prevBG = c.bg
	}

	// Closing separator: prevBG as fg, transparent background.
	if !prevBG.IsEmpty() {
		sb.WriteString("\x1b[0m") // reset to transparent background
		sb.WriteString(toFGSeq(prevBG))
		sb.WriteString(symbol)
		sb.WriteString("\x1b[0m") // full reset
	}

	return sb.String()
}

func (t *TmuxWindowList) renderPlain(windows []tmuxWindow) string {
	parts := make([]string, 0, len(windows))

	for _, win := range windows {
		marker := " "
		if win.Active {
			marker = "*"
		}
		parts = append(parts, fmt.Sprintf("[%s%s:%s]", marker, win.Index, win.Name))
	}

	return strings.Join(parts, " ")
}
