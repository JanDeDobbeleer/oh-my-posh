package shell

import (
	_ "embed"

	"fmt"
	"strings"
)

//go:embed scripts/omp.fish
var fishInit string

func (f Feature) Fish() Code {
	switch f {
	case Transient:
		return "set --global _omp_transient_prompt 1"
	case FTCSMarks:
		return "set --global _omp_ftcs_marks 1"
	case PromptMark:
		return "set --global _omp_prompt_mark 1"
	case Tooltips:
		return "enable_poshtooltips"
	case Upgrade:
		return unixUpgrade
	case Notice:
		return unixNotice
	case RPrompt, PoshGit, Azure, LineError, Jobs, CursorPositioning:
		fallthrough
	default:
		return ""
	}
}

func quoteFishStr(str string) string {
	if len(str) == 0 {
		return "''"
	}
	needQuoting := false
	var b strings.Builder
	for _, r := range str {
		normal := false
		switch r {
		case ';', '"', '(', ')', '[', ']', '{', '}', '$', '|', '&', '>', '<', ' ', '#', '~', '*', '?', '=':
			b.WriteRune(r)
		case '\\', '\'':
			b.WriteByte('\\')
			b.WriteRune(r)
		default:
			b.WriteRune(r)
			normal = true
		}
		if !normal {
			needQuoting = true
		}
	}
	// single quotes are used when the string contains any special characters
	if needQuoting {
		return fmt.Sprintf("'%s'", b.String())
	}
	return b.String()
}
