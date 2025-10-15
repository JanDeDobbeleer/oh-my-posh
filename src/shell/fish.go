package shell

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed scripts/omp.fish
var fishInit string

func (f Features) Fish() Code {
	switch f {
	case Transient:
		return "set --global _omp_transient_prompt 1"
	case FTCSMarks:
		return "set --global _omp_ftcs_marks 1"
	case PromptMark:
		return "set --global _omp_prompt_mark 1"
	case Tooltips:
		return "enable_poshtooltips"
	case Upgrade:
		return unixUpgrade
	case Notice:
		return unixNotice
	case RPrompt, PoshGit, Azure, LineError, Jobs, CursorPositioning, Async:
		fallthrough
	default:
		return ""
	}
}

func quoteFishStr(str string) string {
	if str == "" {
		return "''"
	}

	return fmt.Sprintf("'%s'", strings.NewReplacer(`\`, `\\`, "'", `\'`).Replace(str))
}
