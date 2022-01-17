package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTextSegment(t *testing.T) {
	cases := []struct {
		Case             string
		ExpectedString   string
		Text             string
		ExpectedDisabled bool
	}{
		{Case: "standard text", ExpectedString: "hello", Text: "hello"},
		{Case: "template text with env var", ExpectedString: "hello world", Text: "{{ .Env.HELLO }} world"},
		{Case: "template text with shell name", ExpectedString: "hello world from terminal", Text: "{{ .Env.HELLO }} world from {{ .Shell }}"},
		{Case: "template text with folder", ExpectedString: "hello world in posh", Text: "{{ .Env.HELLO }} world in {{ .Folder }}"},
		{Case: "template text with user", ExpectedString: "hello Posh", Text: "{{ .Env.HELLO }} {{ .UserName }}"},
		{Case: "empty text", Text: "", ExpectedDisabled: true},
		{Case: "empty template result", Text: "{{ .Env.WORLD }}", ExpectedDisabled: true},
	}

	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.On("getcwd").Return("/usr/home/posh")
		env.On("homeDir").Return("/usr/home")
		env.On("getPathSeperator").Return("/")
		env.On("isRunningAsRoot").Return(true)
		env.On("getShellName").Return("terminal")
		env.On("getenv", "HELLO").Return("hello")
		env.On("getenv", "WORLD").Return("")
		env.On("getCurrentUser").Return("Posh")
		env.On("getHostName").Return("MyHost", nil)
		env.onTemplate()
		txt := &text{
			env: env,
			props: properties{
				TextProperty: tc.Text,
			},
		}
		assert.Equal(t, tc.ExpectedDisabled, !txt.enabled(), tc.Case)
		assert.Equal(t, tc.ExpectedString, txt.string(), tc.Case)
	}
}
