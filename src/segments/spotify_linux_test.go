//go:build !darwin && !windows

package segments

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"

	"github.com/stretchr/testify/assert"
)

func TestSpotifyLinux(t *testing.T) {
	cases := []struct {
		Case            string
		ExpectedString  string
		ExpectedEnabled bool
		Title           string
		Artist          string
		Track           string
		Error           error
	}{
		{
			Case:            "Playing",
			ExpectedString:  "\ue602 Snow in Stockholm - Sarah, the Illstrumentalist",
			ExpectedEnabled: true,
			Title:           "Snow in Stockholm - Sarah, the Illstrumentalist",
			Artist:          "Sarah, the Illstrumentalist",
			Track:           "Snow in Stockholm",
		},
		{
			Case:            "Stopped",
			ExpectedEnabled: false,
			Title:           "Spotify - Web Player",
		},
	}
	for _, tc := range cases {
		env := new(mock.MockedEnvironment)

		env.On("RunShellCommand", "bash", "dbus-send --print-reply --dest=org.mpris.MediaPlayer2.spotify /org/mpris/MediaPlayer2 org.freedesktop.DBus.Properties.Get string:org.mpris.MediaPlayer2.Player string:Metadata").Return(tc.Title, tc.Artist)

		s := &Spotify{
			env: env,
			props: properties.Map{
				Command: "echo hi",
			},
		}
		enable := s.Enable()
		assert.True(t, enable)
		if tc.ExpectedEnabled {
			assert.True(t, s.Enable())
			assert.Equal(t, tc.ExpectedString, renderTemplate(env, s.Template(), s))
		}
	}
}
