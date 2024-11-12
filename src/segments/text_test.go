package segments

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/template"

	"github.com/stretchr/testify/assert"
)

func TestTextSegment(t *testing.T) {
	cases := []struct {
		Case             string
		ExpectedString   string
		Template         string
		ExpectedDisabled bool
	}{
		{Case: "standard text", ExpectedString: "hello", Template: "hello"},
		{Case: "template text with env var", ExpectedString: "hello world", Template: "{{ .Env.HELLO }} world"},
		{Case: "template text with shell name", ExpectedString: "hello world from terminal", Template: "{{ .Env.HELLO }} world from {{ .Shell }}"},
		{Case: "template text with folder", ExpectedString: "hello world in posh", Template: "{{ .Env.HELLO }} world in {{ .Folder }}"},
		{Case: "template text with user", ExpectedString: "hello Posh", Template: "{{ .Env.HELLO }} {{ .UserName }}"},
		{Case: "empty text", Template: "", ExpectedDisabled: true},
		{Case: "empty template result", Template: "{{ .Env.WORLD }}", ExpectedDisabled: true},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("PathSeparator").Return("/")
		env.On("Getenv", "HELLO").Return("hello")
		env.On("Getenv", "WORLD").Return("")

		txt := &Text{}
		txt.Init(properties.Map{}, env)

		template.Cache = &cache.Template{
			UserName: "Posh",
			HostName: "MyHost",
			Shell:    "terminal",
			Root:     true,
			Folder:   "posh",
		}

		assert.Equal(t, tc.ExpectedString, renderTemplate(env, tc.Template, txt), tc.Case)
	}
}
