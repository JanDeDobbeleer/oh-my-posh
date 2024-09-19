package shell

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed scripts/omp.tcsh
var tcshInit string

func (f Feature) Tcsh() Code {
	switch f {
	case Upgrade:
		return `"$_omp_executable" upgrade;`
	case Notice:
		return `"$_omp_executable" notice;`
	case PromptMark, RPrompt, PoshGit, Azure, LineError, Jobs, Tooltips, Transient, FTCSMarks, CursorPositioning:
		fallthrough
	default:
		return ""
	}
}

func quoteCshStr(str string) string {
	if len(str) == 0 {
		return "''"
	}

	// An non-working edge case: there is no way to preserve a newline ('\n') in command substitution.
	// Therefore, we can only get a limited string without newlines for "eval".
	return fmt.Sprintf("'%s'", strings.NewReplacer("'", `'"'"'`,
		"!", `\!`).Replace(str))
}
