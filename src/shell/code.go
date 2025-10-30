package shell

import (
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/text"
)

type Code string

const (
	unixFTCSMarks         Code = "_omp_ftcs_marks=1"
	unixCursorPositioning Code = "_omp_cursor_positioning=1"
	unixUpgrade           Code = `"$_omp_executable" upgrade --auto`
	unixNotice            Code = `"$_omp_executable" notice`
)

func (c Code) Indent(spaces int) Code {
	return Code(strings.Repeat(" ", spaces) + string(c))
}

type Lines []Code

func (l Lines) String(script string) string {
	builder := text.NewBuilder()

	builder.WriteString(script)
	builder.WriteString("\n")

	for i, line := range l {
		builder.WriteString(string(line))

		// add newline if not last line
		if i < len(l)-1 {
			builder.WriteString("\n")
		}
	}

	return builder.String()
}
