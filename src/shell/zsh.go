package shell

import (
	_ "embed"
)

//go:embed scripts/omp.zsh
var zshInit string

const (
	unixUpgrade = "$_omp_executable upgrade"
	unixNotice  = "$_omp_executable notice"
)

func (f Feature) Zsh() Code {
	switch f {
	case CursorPositioning:
		return "_omp_cursor_positioning=1"
	case Tooltips:
		return "enable_poshtooltips"
	case Transient:
		return "_omp_create_widget zle-line-init _omp_zle-line-init"
	case FTCSMarks:
		return "_omp_ftcs_marks=1"
	case Upgrade:
		return unixUpgrade
	case Notice:
		return unixNotice
	case PromptMark, RPrompt, PoshGit, Azure, LineError, Jobs:
		fallthrough
	default:
		return ""
	}
}
