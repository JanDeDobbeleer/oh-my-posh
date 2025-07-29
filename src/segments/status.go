package segments

import (
	"strconv"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
)

const (
	StatusTemplate  properties.Property = "status_template"
	StatusSeparator properties.Property = "status_separator"
)

type Status struct {
	base

	String  string
	Meaning string
	Error   bool
}

func (s *Status) Template() string {
	return " {{ .String }} "
}

func (s *Status) Enabled() bool {
	status, pipeStatus := s.env.StatusCodes()

	s.String = s.formatStatus(status, pipeStatus)
	// Deprecated: Use {{ reason .Code }} instead
	s.Meaning = template.GetReasonFromStatus(status)

	if s.props.GetBool(properties.AlwaysEnabled, false) {
		return true
	}

	return s.Error
}

func (s *Status) formatStatus(status int, pipeStatus string) string {
	statusTemplate := s.props.GetString(StatusTemplate, "{{ .Code }}")

	if status != 0 {
		s.Error = true
	}

	if pipeStatus == "" {
		tmpl := template.New(statusTemplate, s)
		defer tmpl.Release()

		if text, err := tmpl.Render(); err == nil {
			return text
		}

		return strconv.Itoa(status)
	}

	StatusSeparator := s.props.GetString(StatusSeparator, "|")

	var builder strings.Builder

	// use an anaonymous struct to avoid
	// confusion with the template context
	// that already has a .Code global property
	var context struct {
		Code int
	}

	splitted := strings.Split(pipeStatus, " ")
	for i, codeStr := range splitted {
		write := func(text string) {
			if i > 0 {
				builder.WriteString(StatusSeparator)
			}
			builder.WriteString(text)
		}

		code, err := strconv.Atoi(codeStr)
		if err != nil {
			write(codeStr)
			continue
		}

		if code != 0 {
			s.Error = true
		}

		context.Code = code

		tmpl := template.New(statusTemplate, context)
		defer tmpl.Release()

		text, err := tmpl.Render()
		if err != nil {
			write(codeStr)
			continue
		}

		write(text)
	}

	return builder.String()
}
