package shell

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed scripts/omp.bash
var bashInit string

func (f Features) Bash() Code {
	switch f {
	case CursorPositioning:
		return unixCursorPositioning
	case FTCSMarks:
		return unixFTCSMarks
	case Upgrade:
		return unixUpgrade
	case Notice:
		return unixNotice
	case RPrompt:
		if !bashBLEsession {
			return ""
		}

		return `bleopt prompt_rps1='$(
	"$_omp_executable" print right \
		--save-cache \
		--shell=bash \
		--shell-version="$BASH_VERSION" \
		--status="$_omp_status" \
		--pipestatus="${_omp_pipestatus[*]}" \
		--no-status="$_omp_no_status" \
		--execution-time="$_omp_execution_time" \
		--stack-count="$_omp_stack_count" \
		--terminal-width="${COLUMNS-0}" \
		--escape=false
)'`
	case Transient:
		if !bashBLEsession {
			return ""
		}

		return `bleopt prompt_ps1_transient=always
bleopt prompt_ps1_final='$(
    "$_omp_executable" print transient \
        --shell=bash \
        --shell-version="$BASH_VERSION" \
        --escape=false
)'`
	case PromptMark, PoshGit, Azure, LineError, Jobs, Tooltips, Async:
		fallthrough
	default:
		return ""
	}
}

func QuotePosixStr(str string) string {
	if str == "" {
		return "''"
	}

	return fmt.Sprintf("$'%s'", strings.NewReplacer(`\`, `\\`, "'", `\'`).Replace(str))
}
