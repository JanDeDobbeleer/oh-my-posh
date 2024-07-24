package shell

import (
	_ "embed"
)

//go:embed scripts/omp.py
var xonshInit string

func (f Feature) Xonsh() Code {
	switch f {
	case Upgrade:
		return "@($POSH_EXECUTABLE) upgrade"
	case Notice:
		return "@($POSH_EXECUTABLE) notice"
	case PromptMark, RPrompt, PoshGit, Azure, LineError, Jobs, Tooltips, Transient, CursorPositioning, FTCSMarks:
		fallthrough
	default:
		return ""
	}
}
