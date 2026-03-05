package prompt

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/config"
)

// TmuxStatusLeft renders the tmux status-left section from the config's tmux.status_left blocks.
func (e *Engine) TmuxStatusLeft() string {
	if e.Config.Tmux == nil {
		return ""
	}

	return ansiToTmux(e.renderTmuxSection(e.Config.Tmux.StatusLeft.Blocks))
}

// TmuxStatusRight renders the tmux status-right section from the config's tmux.status_right blocks.
func (e *Engine) TmuxStatusRight() string {
	if e.Config.Tmux == nil {
		return ""
	}

	return ansiToTmux(e.renderTmuxSection(e.Config.Tmux.StatusRight.Blocks))
}

// renderTmuxSection renders a slice of blocks using the existing block rendering pipeline
// and returns the concatenated result. Shell integration sequences are intentionally
// skipped here since tmux status bars do not use OSC 133 marks.
func (e *Engine) renderTmuxSection(blocks []*config.Block) string {
	if len(blocks) == 0 {
		return ""
	}

	cycle = &e.Config.Cycle

	for _, block := range blocks {
		e.renderBlock(block, true)
	}

	return e.string()
}

var ansiSGRRe = regexp.MustCompile(`\x1b\[([0-9;]*)m`)

// ansiToTmux converts ANSI SGR escape sequences to tmux format strings.
// Tmux strips ESC (0x1b) bytes from #(command) output in status bars, so raw
// ANSI sequences must be translated to tmux's native #[fg=...,bg=...] syntax.
func ansiToTmux(s string) string {
	return ansiSGRRe.ReplaceAllStringFunc(s, func(match string) string {
		sub := ansiSGRRe.FindStringSubmatch(match)
		if len(sub) < 2 {
			return ""
		}
		return sgrToTmuxAttr(sub[1])
	})
}

// sgrToTmuxAttr converts a single SGR parameter string (the part between ESC[ and m)
// into a tmux #[...] style directive.
func sgrToTmuxAttr(params string) string {
	if params == "" || params == "0" {
		return "#[default]"
	}

	parts := strings.Split(params, ";")
	var attrs []string

	i := 0
	for i < len(parts) {
		switch parts[i] {
		case "0":
			attrs = append(attrs, "default")
			i++
		case "1":
			attrs = append(attrs, "bold")
			i++
		case "2":
			attrs = append(attrs, "dim")
			i++
		case "3":
			attrs = append(attrs, "italics")
			i++
		case "4":
			attrs = append(attrs, "underscore")
			i++
		case "5":
			attrs = append(attrs, "blink")
			i++
		case "7":
			attrs = append(attrs, "reverse")
			i++
		case "9":
			attrs = append(attrs, "strikethrough")
			i++
		case "22":
			attrs = append(attrs, "nobold")
			i++
		case "23":
			attrs = append(attrs, "noitalics")
			i++
		case "24":
			attrs = append(attrs, "nounderscore")
			i++
		case "25":
			attrs = append(attrs, "noblink")
			i++
		case "27":
			attrs = append(attrs, "noreverse")
			i++
		case "29":
			attrs = append(attrs, "nostrikethrough")
			i++
		case "39":
			attrs = append(attrs, "fg=default")
			i++
		case "49":
			attrs = append(attrs, "bg=default")
			i++
		case "53":
			attrs = append(attrs, "overline")
			i++
		case "55":
			attrs = append(attrs, "nooverline")
			i++
		case "38":
			if i+4 < len(parts) && parts[i+1] == "2" {
				// True color: 38;2;R;G;B
				r, _ := strconv.ParseUint(parts[i+2], 10, 8)
				g, _ := strconv.ParseUint(parts[i+3], 10, 8)
				b, _ := strconv.ParseUint(parts[i+4], 10, 8)
				attrs = append(attrs, fmt.Sprintf("fg=#%02x%02x%02x", r, g, b))
				i += 5
			} else if i+2 < len(parts) && parts[i+1] == "5" {
				// 256-color: 38;5;N
				attrs = append(attrs, "fg=colour"+parts[i+2])
				i += 3
			} else {
				i++
			}
		case "48":
			if i+4 < len(parts) && parts[i+1] == "2" {
				// True color: 48;2;R;G;B
				r, _ := strconv.ParseUint(parts[i+2], 10, 8)
				g, _ := strconv.ParseUint(parts[i+3], 10, 8)
				b, _ := strconv.ParseUint(parts[i+4], 10, 8)
				attrs = append(attrs, fmt.Sprintf("bg=#%02x%02x%02x", r, g, b))
				i += 5
			} else if i+2 < len(parts) && parts[i+1] == "5" {
				// 256-color: 48;5;N
				attrs = append(attrs, "bg=colour"+parts[i+2])
				i += 3
			} else {
				i++
			}
		default:
			val, err := strconv.ParseUint(parts[i], 10, 64)
			if err != nil {
				i++
				continue
			}
			names16 := []string{"black", "red", "green", "yellow", "blue", "magenta", "cyan", "white"}
			switch {
			case val >= 30 && val <= 37:
				attrs = append(attrs, "fg="+names16[val-30])
			case val >= 40 && val <= 47:
				attrs = append(attrs, "bg="+names16[val-40])
			case val >= 90 && val <= 97:
				attrs = append(attrs, "fg=bright"+names16[val-90])
			case val >= 100 && val <= 107:
				attrs = append(attrs, "bg=bright"+names16[val-100])
			}
			i++
		}
	}

	if len(attrs) == 0 {
		return ""
	}
	return "#[" + strings.Join(attrs, ",") + "]"
}
