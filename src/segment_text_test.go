package main

import (
	"oh-my-posh/environment"
	"oh-my-posh/mock"
	"oh-my-posh/properties"
	"testing"

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
		env := new(mock.MockedEnvironment)
		env.On("PathSeperator").Return("/")
		env.On("TemplateCache").Return(&environment.TemplateCache{
			UserName: "Posh",
			Env: map[string]string{
				"HELLO": "hello",
				"WORLD": "",
			},
			HostName: "MyHost",
			Shell:    "terminal",
			Root:     true,
			Folder:   "posh",
		})
		txt := &text{
			env: env,
			props: properties.Map{
				properties.SegmentTemplate: tc.Template,
			},
		}
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, tc.Template, txt), tc.Case)
	}
}
