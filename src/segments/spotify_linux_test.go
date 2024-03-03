//go:build linux && !darwin && !windows

package segments

import (
	"errors"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"

	"github.com/stretchr/testify/assert"
)

func TestSpotifyLinux(t *testing.T) {
	cases := []struct {
		Error    error
		Expected string
		Status   string
		Artist   string
		Track    string
		Running  bool
	}{
		{Running: false, Expected: ""},
		{Running: false, Expected: "", Error: errors.New("oops")},
		{Running: true, Expected: "\uF8E3 Candlemass - Spellbreaker", Status: "paused", Artist: "Candlemass", Track: "Spellbreaker"},
		{Running: true, Expected: "\uE602 Candlemass - Spellbreaker", Status: "playing", Artist: "Candlemass", Track: "Spellbreaker"},
	}
	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("IsWsl").Return(false)

		dbusCMD := "dbus-send --print-reply --dest=org.mpris.MediaPlayer2.spotify /org/mpris/MediaPlayer2 org.freedesktop.DBus.Properties.Get string:org.mpris.MediaPlayer2.Player"
		env.On("RunShellCommand", shell.BASH, dbusCMD+" string:PlaybackStatus | awk -F '\"' '/string/ {print tolower($2)}'").Return(tc.Status, nil)
		env.On("RunShellCommand", shell.BASH, dbusCMD+" string:Metadata | awk -F '\"' 'BEGIN {RS=\"entry\"}; /'xesam:artist'/ {a=$4} END {print a}'").Return(tc.Artist, nil)
		env.On("RunShellCommand", shell.BASH, dbusCMD+" string:Metadata | awk -F '\"' 'BEGIN {RS=\"entry\"}; /'xesam:title'/ {t=$4} END {print t}'").Return(tc.Track, nil)

		s := &Spotify{}
		s.Init(properties.Map{}, env)

		enable := s.Enabled()
		assert.True(t, enable)
		if tc.Running {
			assert.True(t, s.Enabled())
			assert.Equal(t, tc.Expected, renderTemplate(env, s.Template(), s))
		}
	}
}
