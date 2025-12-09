package segments

import (
	"strconv"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
	"github.com/jandedobbeleer/oh-my-posh/src/text"
)

const (
	StatusTemplate  options.Option = "status_template"
	StatusSeparator options.Option = "status_separator"
)

type Status struct {
	Base

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

	if s.options.Bool(options.AlwaysEnabled, false) {
		return true
	}

	return s.Error
}

func (s *Status) formatStatus(status int, pipeStatus string) string {
	statusTemplate := s.options.String(StatusTemplate, "{{ .Code }}")

	if status != 0 {
		s.Error = true
	}

	if pipeStatus == "" {
		if txt, err := template.Render(statusTemplate, s); err == nil {
			return txt
		}

		return strconv.Itoa(status)
	}

	StatusSeparator := s.options.String(StatusSeparator, "|")

	builder := text.NewBuilder()

	// use an anaonymous struct to avoid
	// confusion with the template context
	// that already has a .Code global property
	var context struct {
		Code int
	}

	splitted := strings.Split(pipeStatus, " ")
	for i, codeStr := range splitted {
		write := func(txt string) {
			if i > 0 {
				builder.WriteString(StatusSeparator)
			}
			builder.WriteString(txt)
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

		txt, err := template.Render(statusTemplate, context)
		if err != nil {
			write(codeStr)
			continue
		}

		write(txt)
	}

	return builder.String()
}
