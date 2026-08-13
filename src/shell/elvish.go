package shell

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed scripts/omp.elv
var elvishInit string

func quoteElvishStr(str string) string {
	if str == "" {
		return "''"
	}

	return fmt.Sprintf("'%s'", strings.ReplaceAll(str, "'", "''"))
}

func (f Features) Elvish() Code {
	switch f {
	case Upgrade:
		return "$_omp_executable upgrade --auto"
	case Notice:
		return "$_omp_executable notice"
	case PromptMark, RPrompt, PoshGit, Azure, LineError, Jobs, CursorPositioning, Tooltips, Transient, FTCSMarks, Async, Streaming, KeyHandlers, VIMode:
		fallthrough
	default:
		return ""
	}
}
