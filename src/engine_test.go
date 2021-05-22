package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCanWriteRPrompt(t *testing.T) {
	cases := []struct {
		Case               string
		Expected           bool
		TerminalWidth      int
		TerminalWidthError error
		PromptLength       int
		RPromptLength      int
	}{
		{Case: "Width Error", Expected: true, TerminalWidthError: errors.New("burp")},
		{Case: "Terminal > Prompt enabled", Expected: true, TerminalWidth: 200, PromptLength: 100, RPromptLength: 10},
		{Case: "Terminal > Prompt enabled edge", Expected: true, TerminalWidth: 200, PromptLength: 100, RPromptLength: 70},
		{Case: "Terminal > Prompt disabled no breathing", Expected: false, TerminalWidth: 200, PromptLength: 100, RPromptLength: 71},
		{Case: "Prompt > Terminal enabled", Expected: true, TerminalWidth: 200, PromptLength: 300, RPromptLength: 70},
		{Case: "Prompt > Terminal disabled no breathing", Expected: true, TerminalWidth: 200, PromptLength: 300, RPromptLength: 80},
		{Case: "Prompt > Terminal disabled no room", Expected: true, TerminalWidth: 200, PromptLength: 400, RPromptLength: 80},
	}

	for _, tc := range cases {
		env := new(MockedEnvironment)
		env.On("getTerminalWidth", nil).Return(tc.TerminalWidth, tc.TerminalWidthError)
		ansi := &ansiUtils{}
		ansi.init(plain)
		engine := &engine{
			env:  env,
			ansi: ansi,
		}
		engine.rprompt = strings.Repeat("x", tc.RPromptLength)
		engine.console.WriteString(strings.Repeat("x", tc.PromptLength))
		got := engine.canWriteRPrompt()
		assert.Equal(t, tc.Expected, got)
	}
}
