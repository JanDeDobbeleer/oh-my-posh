package shell

import (
	_ "embed"
	"strings"
)

//go:embed scripts/omp.lua
var cmdInit string

func (f Feature) Cmd() Code {
	switch f {
	case Transient:
		return "transient_enabled = true"
	case RPrompt:
		return "rprompt_enabled = true"
	case FTCSMarks:
		return "ftcs_marks_enabled = true"
	case Tooltips:
		return "enable_tooltips()"
	case Upgrade:
		return `os.execute(string.format('"%s" upgrade', omp_executable))`
	case Notice:
		return `os.execute(string.format('"%s" notice', omp_executable))`
	case PromptMark, PoshGit, Azure, LineError, Jobs, CursorPositioning:
		fallthrough
	default:
		return ""
	}
}

func escapeLuaStr(str string) string {
	if len(str) == 0 {
		return str
	}
	// We only replace a minimal set of special characters with corresponding escape sequences, without adding surrounding quotes.
	// That way the result can be later quoted with either single or double quotes in a Lua script.
	return strings.NewReplacer(
		`\`, `\\`,
		"'", `\'`,
		`"`, `\"`,
		"\n", `\n`,
		"\r", `\r`,
	).Replace(str)
}
