package segments

import (
	"strconv"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
	"github.com/jandedobbeleer/oh-my-posh/src/text"
)

type Status struct {
	SegmentBase

	// Configuration fields with defaults
	StatusTemplate  string `json:"status_template,omitempty" toml:"status_template,omitempty" yaml:"status_template,omitempty" default:"{{ .Code }}"`
	StatusSeparator string `json:"status_separator,omitempty" toml:"status_separator,omitempty" yaml:"status_separator,omitempty" default:"|"`
	AlwaysEnabled   bool   `json:"always_enabled,omitempty" toml:"always_enabled,omitempty" yaml:"always_enabled,omitempty"`

	// Runtime state (not serialized)
	String  string `json:"-"`
	Meaning string `json:"-"`
	Error   bool   `json:"-"`
	Code    int    `json:"-"`
}

func (s *Status) Template() string {
	return " {{ .String }} "
}

// Init satisfies the SegmentWriter interface (ignores props for typed segments)
func (s *Status) Init(_ properties.Properties, env runtime.Environment) {
	s.SegmentBase.Init(env)
}

// IsTypedSegment marks this as a typed segment
func (s *Status) IsTypedSegment() {}

func (s *Status) Enabled() bool {
	status, pipeStatus := s.Env().StatusCodes()

	s.Code = status
	s.String = s.formatStatus(status, pipeStatus)
	// Deprecated: Use {{ reason .Code }} instead
	s.Meaning = template.GetReasonFromStatus(status)

	if s.AlwaysEnabled {
		return true
	}

	return s.Error
}

func (s *Status) formatStatus(status int, pipeStatus string) string {
	statusTemplate := s.StatusTemplate

	if status != 0 {
		s.Error = true
	}

	if pipeStatus == "" {
		s.Code = status
		if txt, err := template.Render(statusTemplate, s); err == nil {
			return txt
		}

		return strconv.Itoa(status)
	}

	statusSeparator := s.StatusSeparator

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
				builder.WriteString(statusSeparator)
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
