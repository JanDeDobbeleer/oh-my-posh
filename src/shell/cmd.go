package shell

import (
	_ "embed"

	"fmt"
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
	case Tooltips:
		return "enable_tooltips()"
	case Upgrade:
		return "os.execute(string.format('%s upgrade', omp_exe()))"
	case Notice:
		return "os.execute(string.format('%s notice', omp_exe()))"
	case PromptMark, PoshGit, Azure, LineError, Jobs, FTCSMarks, CursorPositioning:
		fallthrough
	default:
		return ""
	}
}

func quoteLuaStr(str string) string {
	if len(str) == 0 {
		return "''"
	}

	return fmt.Sprintf("'%s'", strings.NewReplacer(`\`, `\\`, `'`, `\'`).Replace(str))
}
