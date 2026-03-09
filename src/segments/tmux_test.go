package segments

import (
	"errors"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"

	"github.com/stretchr/testify/assert"
)

const tmuxListWindowsFmt = "#{window_index}\t#{window_name}\t#{window_active}"

func TestTmuxEnabled(t *testing.T) {
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
			// $TMUX third field (index 2) is used as last-resort identifier
			SessionName: "0",
		},
		{
			Case:       "tmux command fails, TMUX env malformed",
			CommandErr: errors.New("not in tmux"),
			TmuxEnv:    "bad-value",
			Expected:   false,
		},
		{
			Case:       "tmux command fails, TMUX env empty — not in tmux",
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

		seg := &Tmux{}
		seg.Init(options.Map{}, env)

		result := seg.Enabled()

		assert.Equal(t, tc.Expected, result, tc.Case)
		if tc.Expected {
			assert.Equal(t, tc.SessionName, seg.SessionName, tc.Case)
		}
	}
}

func TestTmuxNoWindowsByDefault(t *testing.T) {
	env := new(mock.Environment)
	env.On("RunCommand", "tmux", []string{"display-message", "-p", "#S"}).
		Return("work\n", nil)

	seg := &Tmux{}
	seg.Init(options.Map{}, env)

	enabled := seg.Enabled()
	assert.True(t, enabled)
	assert.Nil(t, seg.Windows, "Windows should be nil when fetch_windows is false")
}

func TestTmuxFetchWindows(t *testing.T) {
	env := new(mock.Environment)
	env.On("RunCommand", "tmux", []string{"display-message", "-p", "#S"}).
		Return("work\n", nil)
	env.On("RunCommand", "tmux", []string{"list-windows", "-F", tmuxListWindowsFmt}).
		Return("0\tbash\t1\n1\tnvim\t0\n", nil)

	seg := &Tmux{}
	seg.Init(options.Map{fetchWindows: true}, env)

	enabled := seg.Enabled()
	assert.True(t, enabled)
	assert.Len(t, seg.Windows, 2)
	assert.Equal(t, "0", seg.Windows[0].Index)
	assert.Equal(t, "bash", seg.Windows[0].Name)
	assert.True(t, seg.Windows[0].Active)
	assert.Equal(t, "1", seg.Windows[1].Index)
	assert.Equal(t, "nvim", seg.Windows[1].Name)
	assert.False(t, seg.Windows[1].Active)
}

func TestTmuxFetchWindowsCommandFails(t *testing.T) {
	env := new(mock.Environment)
	env.On("RunCommand", "tmux", []string{"display-message", "-p", "#S"}).
		Return("work\n", nil)
	env.On("RunCommand", "tmux", []string{"list-windows", "-F", tmuxListWindowsFmt}).
		Return("", errors.New("not in tmux"))

	seg := &Tmux{}
	seg.Init(options.Map{fetchWindows: true}, env)

	enabled := seg.Enabled()
	// Segment is still enabled — session name was fetched successfully.
	assert.True(t, enabled)
	assert.Nil(t, seg.Windows)
}

func TestTmuxFetchWindowsEmptyOutput(t *testing.T) {
	env := new(mock.Environment)
	env.On("RunCommand", "tmux", []string{"display-message", "-p", "#S"}).
		Return("work\n", nil)
	env.On("RunCommand", "tmux", []string{"list-windows", "-F", tmuxListWindowsFmt}).
		Return("", nil)

	seg := &Tmux{}
	seg.Init(options.Map{fetchWindows: true}, env)

	enabled := seg.Enabled()
	assert.True(t, enabled)
	assert.Empty(t, seg.Windows)
}

func TestTmuxParseWindows(t *testing.T) {
	seg := &Tmux{}

	windows := seg.parseWindows("0\tbash\t1\n1\tnvim\t0\n2\thtop\t0")
	assert.Len(t, windows, 3)
	assert.Equal(t, "0", windows[0].Index)
	assert.Equal(t, "bash", windows[0].Name)
	assert.True(t, windows[0].Active)
	assert.Equal(t, "1", windows[1].Index)
	assert.Equal(t, "nvim", windows[1].Name)
	assert.False(t, windows[1].Active)
	assert.Equal(t, "2", windows[2].Index)
	assert.Equal(t, "htop", windows[2].Name)
	assert.False(t, windows[2].Active)
}

func TestTmuxParseWindowsSkipsMalformedLines(t *testing.T) {
	seg := &Tmux{}

	windows := seg.parseWindows("0\tbash\t1\nbad-line\n1\tnvim\t0")
	assert.Len(t, windows, 2)
	assert.Equal(t, "bash", windows[0].Name)
	assert.Equal(t, "nvim", windows[1].Name)
}

func TestTmuxTemplate(t *testing.T) {
	seg := &Tmux{}
	assert.Contains(t, seg.Template(), "{{ .SessionName }}")
	assert.Contains(t, seg.Template(), ".Windows")
}
