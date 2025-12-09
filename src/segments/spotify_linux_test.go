//go:build linux && !darwin && !windows

package segments

import (
	"errors"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"

	"github.com/stretchr/testify/assert"
)

func TestSpotifyLinux(t *testing.T) {
	cases := []struct {
		Case            string
		Status          string
		Artist          string
		Track           string
		Expected        string
		ExpectedEnabled bool
	}{
		{Case: "no data", ExpectedEnabled: false},
		{Case: "error", ExpectedEnabled: false, Status: "Error.ServiceUnknown"},
		{Case: "paused", ExpectedEnabled: true, Expected: "\uF8E3 Candlemass - Spellbreaker", Status: "paused", Artist: "Candlemass", Track: "Spellbreaker"},
		{Case: "playing", ExpectedEnabled: true, Expected: "\uE602 Candlemass - Spellbreaker", Status: "playing", Artist: "Candlemass", Track: "Spellbreaker"},
	}
	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("IsWsl").Return(false)

		dbusCMD := "dbus-send --print-reply --dest=org.mpris.MediaPlayer2.spotify /org/mpris/MediaPlayer2 org.freedesktop.DBus.options.Get string:org.mpris.MediaPlayer2.Player"
		env.On("RunShellCommand", shell.BASH, dbusCMD+" string:PlaybackStatus | awk -F '\"' '/string/ {print tolower($2)}'").Return(tc.Status)
		env.On("RunShellCommand", shell.BASH, dbusCMD+" string:Metadata | awk -F '\"' 'BEGIN {RS=\"entry\"}; /'xesam:artist'/ {a=$4} END {print a}'").Return(tc.Artist)
		env.On("RunShellCommand", shell.BASH, dbusCMD+" string:Metadata | awk -F '\"' 'BEGIN {RS=\"entry\"}; /'xesam:title'/ {t=$4} END {print t}'").Return(tc.Track)

		s := &Spotify{}
		s.Init(options.Map{}, env)

		got := s.Enabled()

		assert.Equal(t, tc.ExpectedEnabled, got, tc.Case)
		if tc.ExpectedEnabled {
			assert.Equal(t, tc.Expected, renderTemplate(env, s.Template(), s), tc.Case)
		}
	}
}

func TestSpotifyWSL(t *testing.T) {
	cases := []struct {
		Case            string
		Error           error
		Title           string
		Expected        string
		ExpectedEnabled bool
	}{
		{Case: "nothing"},
		{Case: "error", Error: errors.New("oops")},
		{Case: "title", ExpectedEnabled: true, Expected: "\ue602 Xzibit - Crash (ft. Royce Da 5'9\" K.A.A.N.)", Title: `Xzibit - Crash (ft. Royce Da 5'9" K.A.A.N.)`},
	}
	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("IsWsl").Return(true)

		psCommand := `[Console]::OutputEncoding = [System.Text.Encoding]::UTF8; (Get-Process Spotify -ErrorAction SilentlyContinue | Where-Object {$_.MainWindowTitle -ne ""} | Select-Object -First 1).MainWindowTitle` //nolint: lll
		env.On("RunCommand", "powershell.exe", []string{"-NoProfile", "-NonInteractive", "-Command", psCommand}).Return(tc.Title, tc.Error)

		s := &Spotify{}
		s.Init(options.Map{}, env)

		got := s.Enabled()

		assert.Equal(t, tc.ExpectedEnabled, got, tc.Case)
		if tc.ExpectedEnabled {
			assert.Equal(t, tc.Expected, renderTemplate(env, s.Template(), s), tc.Case)
		}
	}
}
