package shell

import (
	_ "embed"
)

//go:embed scripts/omp.elv
var elvishInit string

func (f Feature) Elvish() Code {
	switch f {
	case Upgrade:
		return "$_omp_executable upgrade"
	case Notice:
		return "$_omp_executable notice"
	case PromptMark, RPrompt, PoshGit, Azure, LineError, Jobs, CursorPositioning, Tooltips, Transient, FTCSMarks:
		fallthrough
	default:
		return ""
	}
}
