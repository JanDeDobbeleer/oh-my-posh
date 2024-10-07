package segments

import (
	"strconv"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
)

const (
	StatusTemplate  properties.Property = "status_template"
	StatusSeparator properties.Property = "status_separator"
)

type Status struct {
	props    properties.Properties
	env      runtime.Environment
	template *template.Text
	String   string
	Meaning  string
	Code     int
	Error    bool
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

func (s *Status) Init(props properties.Properties, env runtime.Environment) {
	s.props = props
	s.env = env

	statusTemplate := s.props.GetString(StatusTemplate, "{{ .Code }}")
	s.template = &template.Text{
		Template: statusTemplate,
	}
}

func (s *Status) formatStatus(status int, pipeStatus string) string {
	if status != 0 {
		s.Error = true
	}

	if len(pipeStatus) == 0 {
		s.Code = status
		s.template.Context = s
		if text, err := s.template.Render(); err == nil {
			return text
		}
		return strconv.Itoa(status)
	}

	StatusSeparator := s.props.GetString(StatusSeparator, "|")

	var builder strings.Builder

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

		s.Code = code
		s.template.Context = s
		text, err := s.template.Render()
		if err != nil {
			write(codeStr)
			continue
		}

		write(text)
	}

	return builder.String()
}
