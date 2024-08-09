//go:build linux && !darwin && !windows

package segments

import (
	"errors"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"

	"github.com/stretchr/testify/assert"
)

func TestSpotifyLinux(t *testing.T) {
	cases := []struct {
		Running  bool
		Expected string
		Status   string
		Artist   string
		Track    string
		Error    error
	}{
		{Running: false, Expected: "\uE602  -", Error: errors.New("oops")},
		{Running: true, Expected: "\uE602 Candlemass - Spellbreaker", Status: "playing", Artist: "Candlemass", Track: "Spellbreaker"},
		{Running: true, Expected: "\uE602 Candlemass - Spellbreaker", Status: "paused", Artist: "Candlemass", Track: "Spellbreaker"},
	}
	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("IsWsl").Return(false)
		env.On("RunShellCommand", shell.BASH, "dbus-send --print-reply --dest=org.mpris.MediaPlayer2.spotify /org/mpris/MediaPlayer2 org.freedesktop.DBus.Properties.Get string:org.mpris.MediaPlayer2.Player string:Metadata").Return(tc.Status, nil)
		env.On("RunShellCommand", shell.BASH, "dbus-send --print-reply --dest=org.mpris.MediaPlayer2.spotify /org/mpris/MediaPlayer2 org.freedesktop.DBus.Properties.Get string:org.mpris.MediaPlayer2.Player string:Metadata | awk -F '\"' 'BEGIN {RS=\"entry\"}; /'xesam:artist'/ {a=$4} END {print a}'").Return(tc.Artist, nil)
		env.On("RunShellCommand", shell.BASH, "dbus-send --print-reply --dest=org.mpris.MediaPlayer2.spotify /org/mpris/MediaPlayer2 org.freedesktop.DBus.Properties.Get string:org.mpris.MediaPlayer2.Player string:Metadata | awk -F '\"' 'BEGIN {RS=\"entry\"}; /'xesam:title'/ {t=$4} END {print t}'").Return(tc.Track, nil)

		s := &Spotify{
			env:   env,
			props: properties.Map{},
		}
		enable := s.Enabled()
		assert.True(t, enable)
		if tc.Running {
			assert.True(t, s.Enabled())
			assert.Equal(t, tc.Expected, renderTemplate(env, s.Template(), s))
		}
	}
}
