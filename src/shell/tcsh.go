package shell

import (
	_ "embed"
)

//go:embed scripts/omp.tcsh
var tcshInit string

func (f Feature) Tcsh() Code {
	switch f {
	case Upgrade:
		return "$POSH_COMMAND upgrade;"
	case Notice:
		return "$POSH_COMMAND notice;"
	case PromptMark, RPrompt, PoshGit, Azure, LineError, Jobs, Tooltips, Transient, FTCSMarks, CursorPositioning:
		fallthrough
	default:
		return ""
	}
}
