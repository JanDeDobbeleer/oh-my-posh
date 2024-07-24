package shell

import "strings"

type Code string

func (c Code) Indent(spaces int) Code {
	return Code(strings.Repeat(" ", spaces) + string(c))
}

type Lines []Code

func (l Lines) String(script string) string {
	var builder strings.Builder

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
