package template

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/stretchr/testify/assert"
	testify_ "github.com/stretchr/testify/mock"
)

func TestRoundSeconds(t *testing.T) {
	cases := []struct {
		Case        string
		Expected    string
		Template    string
		ShouldError bool
	}{
		{Case: "int - 1 second", Expected: "1s", Template: "{{ secondsRound 1 }}"},
		{Case: "double - 1 second", Expected: "1s", Template: "{{ secondsRound 1.1 }}"},
		{Case: "int - 1 minute", Expected: "1m", Template: "{{ secondsRound 60 }}"},
		{Case: "int - 2 minutes 30 seconds", Expected: "2m 30s", Template: "{{ secondsRound 150 }}"},
		{Case: "int - 1 day 2 minutes 30 seconds", Expected: "1d 2m 30s", Template: "{{ secondsRound 86550 }}"},
		{Case: "double - 1 day 2 minutes 30 seconds", Expected: "1d 2m 30s", Template: "{{ secondsRound 86550.555 }}"},
		{Case: "int - 1 month 1 day 2 minutes 30 seconds", Expected: "1mo 1d 2m 30s", Template: "{{ secondsRound 2716350 }}"},
		{Case: "int - 1 year 1 month 1 day 2 minutes 30 seconds", Expected: "1y 1mo 1d 2m 30s", Template: "{{ secondsRound 34276350 }}"},
		{Case: "error", Expected: "", Template: "{{ secondsRound foo }}", ShouldError: true},
	}

	env := &mock.Environment{}
	env.On("TemplateCache").Return(&cache.Template{
		Env: make(map[string]string),
	})
	env.On("Error", testify_.Anything)
	env.On("Debug", testify_.Anything)
	env.On("DebugF", testify_.Anything, testify_.Anything).Return(nil)
	for _, tc := range cases {
		tmpl := &Text{
			Template: tc.Template,
			Context:  nil,
			Env:      env,
		}
		text, err := tmpl.Render()
		if tc.ShouldError {
			assert.Error(t, err)
			continue
		}
		assert.Equal(t, tc.Expected, text, tc.Case)
	}
}
