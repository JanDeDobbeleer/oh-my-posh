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
		env.On("getPathSeperator").Return("/")
		env.On("templateCache").Return(&templateCache{
			UserName: "Posh",
			Env: map[string]string{
				"HELLO": "hello",
				"WORLD": "",
			},
			HostName: "MyHost",
			Shell:    "terminal",
			Root:     true,
			Folder:   base("/usr/home/posh", env),
		})
		txt := &text{
			env: env,
			props: properties{
				SegmentTemplate: tc.Text,
			},
		}
		assert.Equal(t, tc.ExpectedDisabled, !txt.enabled(), tc.Case)
		assert.Equal(t, tc.ExpectedString, txt.string(), tc.Case)
	}
}
