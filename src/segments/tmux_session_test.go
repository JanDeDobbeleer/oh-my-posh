package segments

import (
	"errors"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"

	"github.com/stretchr/testify/assert"
)

func TestTmuxSessionEnabled(t *testing.T) {
	cases := []struct {
		Case        string
		CommandOut  string
		CommandErr  error
		TmuxEnv     string
		Expected    bool
		SessionName string
	}{
		{
			Case:        "tmux command succeeds",
			CommandOut:  "mysession\n",
			Expected:    true,
			SessionName: "mysession",
		},
		{
			Case:        "tmux command succeeds with whitespace",
			CommandOut:  "  main  \n",
			Expected:    true,
			SessionName: "main",
		},
		{
			Case:       "tmux command fails, fallback to TMUX env",
			CommandErr: errors.New("not in tmux"),
			TmuxEnv:    "/tmp/tmux-1000/default,12345,0",
			Expected:   true,
			// $TMUX third field (index 2) is used as session name
			SessionName: "0",
		},
		{
			Case:       "tmux command fails, TMUX env malformed",
			CommandErr: errors.New("not in tmux"),
			TmuxEnv:    "bad-value",
			Expected:   false,
		},
		{
			Case:       "tmux command fails, TMUX env empty",
			CommandErr: errors.New("not in tmux"),
			TmuxEnv:    "",
			Expected:   false,
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("RunCommand", "tmux", []string{"display-message", "-p", "#S"}).
			Return(tc.CommandOut, tc.CommandErr)
		env.On("Getenv", "TMUX").Return(tc.TmuxEnv)

		seg := &TmuxSession{}
		seg.Init(options.Map{}, env)

		result := seg.Enabled()

		assert.Equal(t, tc.Expected, result, tc.Case)
		if tc.Expected {
			assert.Equal(t, tc.SessionName, seg.SessionName, tc.Case)
		}
	}
}

func TestTmuxSessionTemplate(t *testing.T) {
	env := new(mock.Environment)
	env.On("RunCommand", "tmux", []string{"display-message", "-p", "#S"}).
		Return("work\n", nil)

	seg := &TmuxSession{}
	seg.Init(options.Map{}, env)
	_ = seg.Enabled()

	assert.Equal(t, "work", seg.SessionName)
	assert.Equal(t, " \ue7a2 {{ .SessionName }} ", seg.Template())
}
