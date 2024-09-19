package shell

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed scripts/omp.xsh
var xonshInit string

func (f Feature) Xonsh() Code {
	switch f {
	case Upgrade:
		return "@(_omp_executable) upgrade"
	case Notice:
		return "@(_omp_executable) notice"
	case PromptMark, RPrompt, PoshGit, Azure, LineError, Jobs, Tooltips, Transient, CursorPositioning, FTCSMarks:
		fallthrough
	default:
		return ""
	}
}

func quotePythonStr(str string) string {
	if len(str) == 0 {
		return "''"
	}

	return fmt.Sprintf("'%s'", strings.NewReplacer(
		"'", `'"'"'`,
		`\`, `\\`,
		"\n", `\n`,
	).Replace(str))
}
