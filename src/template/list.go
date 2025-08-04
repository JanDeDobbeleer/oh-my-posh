package template

import (
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/text"
)

type Logic string

const (
	FirstMatch Logic = "first_match"
	Join       Logic = "join"
)

type List []string

func (l List) Empty() bool {
	return len(l) == 0
}

func (l List) Resolve(context any, defaultValue string, logic Logic) string {
	if l.Empty() {
		return defaultValue
	}

	switch logic {
	case FirstMatch:
		return l.FirstMatch(context, defaultValue)
	case Join:
		fallthrough
	default:
		return l.Join(context)
	}
}

func (l List) Join(context any) string {
	if len(l) == 0 {
		return ""
	}

	buffer := text.NewBuilder()

	for _, tmpl := range l {
		value, err := Render(tmpl, context)
		if err != nil || len(strings.TrimSpace(value)) == 0 {
			continue
		}

		buffer.WriteString(value)
	}

	return buffer.String()
}

func (l List) FirstMatch(context any, defaultValue string) string {
	if len(l) == 0 {
		return defaultValue
	}

	for _, tmpl := range l {
		value, err := Render(tmpl, context)
		if err != nil || len(strings.TrimSpace(value)) == 0 {
			continue
		}
		return value
	}

	return defaultValue
}
