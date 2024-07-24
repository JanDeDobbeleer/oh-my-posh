package shell

import (
	_ "embed"
)

//go:embed scripts/omp.elv
var elvishInit string

func (f Feature) Elvish() Code {
	switch f {
	case Upgrade:
		return unixUpgrade
	case Notice:
		return unixNotice
	case PromptMark, RPrompt, PoshGit, Azure, LineError, Jobs, CursorPositioning, Tooltips, Transient, FTCSMarks:
		fallthrough
	default:
		return ""
	}
}
