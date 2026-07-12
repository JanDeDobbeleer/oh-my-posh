package shell

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed scripts/omp.yash
var yashInit string

// quoteYashStr quotes a string with plain POSIX single quotes.
// Yash only understands $'...' since it implemented POSIX 2024,
// so QuotePosixStr's output breaks on older releases.
func quoteYashStr(str string) string {
	if str == "" {
		return "''"
	}

	return fmt.Sprintf("'%s'", strings.ReplaceAll(str, "'", `'\''`))
}

func (f Features) Yash() Code {
	switch f {
	case FTCSMarks:
		return unixFTCSMarks
	case Upgrade:
		return unixUpgrade
	case Notice:
		return unixNotice
	case RPrompt:
		return `YASH_PS1R='$(
    "$_omp_executable" print right \
        --save-cache \
        --shell=yash \
        --shell-version="$YASH_VERSION" \
        --status="$_omp_status" \
        --no-status="$_omp_no_status" \
        --execution-time="$_omp_execution_time" \
        --job-count="$_omp_job_count" \
        --terminal-width="${COLUMNS:-0}"
)'`
	case Transient, Tooltips, KeyHandlers, CursorPositioning, Async, Streaming, VIMode, LineError, Jobs, Azure, PoshGit, PromptMark:
		fallthrough
	default:
		return ""
	}
}
