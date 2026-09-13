package segments

import (
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
)

type Shell struct {
	Base

	// Name is markup: a mapped_shell_names value is user configuration and
	// may carry <...> anchors, while the detected shell name is escaped.
	Name    template.Markup
	Version string
}

const (
	MappedShellNames options.Option = "mapped_shell_names"
)

func (s *Shell) Template() string {
	return NameTemplate
}

func (s *Shell) Enabled() bool {
	mappedNames := s.options.KeyValueMap(MappedShellNames, make(map[string]string))
	name := s.env.Shell()
	s.Name = template.EscapeMarkup(name)
	s.Version = s.env.Flags().ShellVersion
	for key, val := range mappedNames {
		if strings.EqualFold(name, key) {
			s.Name = template.RawMarkup(val)
			break
		}
	}
	return true
}
