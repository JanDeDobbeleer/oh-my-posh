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
			ExpectedString:  "Sarah, the Illstrumentalist - Snow in Stockholm",
			ExpectedEnabled: true,
			Title:           "Snow in Stockholm - Sarah, the Illstrumentalist",
			Artist:          "Sarah, the Illstrumentalist",
			Track:           "Snow in Stockholm",
		},
		{
			Case:            "Stopped",
			ExpectedEnabled: true,
			ExpectedString:  "-",
		},
	}
	for _, tc := range cases {
		env := new(mock.MockedEnvironment)

		env.On("RunCommand", "dbus-send", []string{
			"--print-reply",
			"--dest=org.mpris.MediaPlayer2.spotify /org/mpris/MediaPlayer2 org.freedesktop.DBus.Properties.Get string:org.mpris.MediaPlayer2.Player string:Metadata",
		}).Return(tc.Track+" "+tc.Artist, nil)

		s := &Spotify{
			env:   env,
			props: properties.Map{},
			MusicPlayer: MusicPlayer{
				Artist: tc.Artist,
				Track:  tc.Track,
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
