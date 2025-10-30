package shell

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed scripts/omp.nu
var nuInit string

func (f Features) Nu() Code {
	switch f {
	case Transient:
		return `$env.TRANSIENT_PROMPT_COMMAND = {|| _omp_get_prompt transient }`
	case Upgrade:
		return "^$_omp_executable upgrade --auto"
	case Notice:
		return "^$_omp_executable notice"
	case PromptMark, RPrompt, PoshGit, Azure, LineError, Jobs, Tooltips, FTCSMarks, CursorPositioning, Async:
		fallthrough
	default:
		return ""
	}
}

func quoteNuStr(str string) string {
	if str == "" {
		return "''"
	}

	return fmt.Sprintf(`"%s"`, strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(str))
}
