package shell

import (
	_ "embed"

	"fmt"
	"strings"
)

//go:embed scripts/omp.bash
var bashInit string

func (f Feature) Bash() Code {
	switch f {
	case CursorPositioning:
		return "_omp_cursor_positioning=1"
	case FTCSMarks:
		return "_omp_ftcs_marks=1"
	case Upgrade:
		return `"$_omp_executable" upgrade`
	case Notice:
		return `"$_omp_executable" notice`
	case PromptMark, RPrompt, PoshGit, Azure, LineError, Jobs, Tooltips, Transient:
		fallthrough
	default:
		return ""
	}
}

func QuotePosixStr(str string) string {
	if len(str) == 0 {
		return "''"
	}

	needQuoting := false
	var b strings.Builder
	for _, r := range str {
		normal := false
		switch r {
		case '!', ';', '"', '(', ')', '[', ']', '{', '}', '$', '|', '&', '>', '<', '`', ' ', '#', '~', '*', '?', '=':
			b.WriteRune(r)
		case '\\', '\'':
			b.WriteByte('\\')
			b.WriteRune(r)
		case '\a':
			b.WriteString(`\a`)
		case '\b':
			b.WriteString(`\b`)
		case '\f':
			b.WriteString(`\f`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		case '\v':
			b.WriteString(`\v`)
		default:
			b.WriteRune(r)
			normal = true
		}
		if !normal {
			needQuoting = true
		}
	}
	// the quoting form $'...' is used for a string contains any special characters
	if needQuoting {
		return fmt.Sprintf("$'%s'", b.String())
	}

	return b.String()
}
