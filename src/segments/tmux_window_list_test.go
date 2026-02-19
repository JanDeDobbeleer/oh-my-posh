package segments

import (
	"errors"
	"strings"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"

	"github.com/stretchr/testify/assert"
)

const tmuxListWindowsFormat = "#{window_index}\t#{window_name}\t#{window_active}"

func TestTmuxWindowListEnabled(t *testing.T) {
	cases := []struct {
		Case       string
		Output     string
		Err        error
		Expected   bool
		WindowsLen int
	}{
		{
			Case:       "two windows, first active",
			Output:     "0\tbash\t1\n1\tnvim\t0\n",
			Expected:   true,
			WindowsLen: 2,
		},
		{
			Case:       "single window active",
			Output:     "0\tzsh\t1\n",
			Expected:   true,
			WindowsLen: 1,
		},
		{
			Case:     "command fails",
			Err:      errors.New("not in tmux"),
			Expected: false,
		},
		{
			Case:     "empty output",
			Output:   "",
			Expected: false,
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("RunCommand", "tmux", []string{"list-windows", "-F", tmuxListWindowsFormat}).
			Return(tc.Output, tc.Err)

		terminal.Plain = true // use plain mode to avoid needing terminal.Colors

		seg := &TmuxWindowList{}
		seg.Init(options.Map{}, env)

		result := seg.Enabled()

		assert.Equal(t, tc.Expected, result, tc.Case)
	}

	terminal.Plain = false // restore
}

func TestTmuxWindowListPlainRendering(t *testing.T) {
	cases := []struct {
		Case       string
		Output     string
		Contains   []string
		NotContain []string
	}{
		{
			Case:     "active window marked with asterisk",
			Output:   "0\tbash\t1\n1\tnvim\t0\n",
			Contains: []string{"*0:bash", " 1:nvim"},
		},
		{
			Case:     "all inactive",
			Output:   "0\tbash\t0\n1\tnvim\t0\n",
			Contains: []string{" 0:bash", " 1:nvim"},
		},
		{
			Case:     "single active window",
			Output:   "3\twork\t1\n",
			Contains: []string{"*3:work"},
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("RunCommand", "tmux", []string{"list-windows", "-F", tmuxListWindowsFormat}).
			Return(tc.Output, nil)

		terminal.Plain = true

		seg := &TmuxWindowList{}
		seg.Init(options.Map{}, env)

		enabled := seg.Enabled()
		assert.True(t, enabled, tc.Case)

		for _, want := range tc.Contains {
			assert.True(t, strings.Contains(seg.RenderedList, want), "%s: expected %q in %q", tc.Case, want, seg.RenderedList)
		}
	}

	terminal.Plain = false
}

func TestTmuxWindowListParseWindows(t *testing.T) {
	seg := &TmuxWindowList{}

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

func TestTmuxWindowListSkipsMalformedLines(t *testing.T) {
	seg := &TmuxWindowList{}

	// Lines with fewer than 3 tab-separated fields should be skipped.
	windows := seg.parseWindows("0\tbash\t1\nbad-line\n1\tnvim\t0")
	assert.Len(t, windows, 2)
	assert.Equal(t, "bash", windows[0].Name)
	assert.Equal(t, "nvim", windows[1].Name)
}

func TestTmuxWindowListTemplate(t *testing.T) {
	seg := &TmuxWindowList{}
	assert.Equal(t, "{{ .RenderedList }}", seg.Template())
}
