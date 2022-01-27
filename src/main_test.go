package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConsoleBackgroundColorTemplate(t *testing.T) {
	cases := []struct {
		Case     string
		Expected string
		Term     string
	}{
		{Case: "Inside vscode", Expected: "#123456", Term: "vscode"},
		{Case: "Outside vscode", Expected: "", Term: "windowsterminal"},
	}

	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.On("TemplateCache").Return(&TemplateCache{
			Env: map[string]string{
				"TERM_PROGRAM": tc.Term,
			},
		})
		color := getConsoleBackgroundColor(env, "{{ if eq \"vscode\" .Env.TERM_PROGRAM }}#123456{{end}}")
		assert.Equal(t, tc.Expected, color, tc.Case)
	}
}
