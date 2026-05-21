package shell

import (
	_ "embed"
	"strings"
)

//go:embed scripts/omp.lua
var cmdInit string

func (f Features) Cmd() Code {
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
		return `os.execute(string.format('"%s" upgrade --auto', omp_executable))`
	case Notice:
		return `if clink.onbeginedit then
    local need_notice = true
    clink.onbeginedit(function()
        if need_notice then
            need_notice = false
            os.execute(string.format('"%s" notice', omp_executable))
        end
    end)
else
    os.execute(string.format('"%s" notice', omp_executable))
end`
	case PromptMark, PoshGit, Azure, LineError, Jobs, CursorPositioning, Async, Streaming, KeyHandlers, VIMode:
		fallthrough
	default:
		return ""
	}
}

func escapeLuaStr(str string) string {
	if str == "" {
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
