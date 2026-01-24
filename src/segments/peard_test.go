package segments

import (
	"fmt"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"

	"github.com/stretchr/testify/assert"
)

func TestPearDesktop(t *testing.T) {
	cases := []struct {
		Case            string
		JSONResponse    string
		ExpectedString  string
		ExpectedEnabled bool
		ExpectedStatus  string
		Port            int
		HTTPError       error
	}{
		{
			Case:            "playing",
			ExpectedEnabled: true,
			ExpectedStatus:  "playing",
			JSONResponse: `{
				"title": "Title",
				"artist": "Artist",
				"isPaused": false,
				"songDuration": 221,
				"elapsedSeconds": 100
			}`,
			ExpectedString: "Playing Artist - Title",
		},
		{
			Case:            "paused",
			ExpectedEnabled: true,
			ExpectedStatus:  "paused",
			JSONResponse: `{
				"title": "Title",
				"artist": "Artist",
				"isPaused": true,
				"songDuration": 221,
				"elapsedSeconds": 100
			}`,
			ExpectedString: "Paused",
		},
		{
			Case:            "http error",
			ExpectedEnabled: false,
			HTTPError:       assert.AnError,
		},
		{
			Case:            "empty response",
			ExpectedEnabled: false,
			JSONResponse:    "",
		},
		{
			Case:            "invalid json",
			ExpectedEnabled: false,
			JSONResponse:    "invalid json",
		},
		{
			Case:            "custom port",
			ExpectedEnabled: true,
			ExpectedStatus:  "playing",
			Port:            12345,
			JSONResponse: `{
				"title": "Title",
				"artist": "Artist",
				"isPaused": false
			}`,
			ExpectedString: "Playing Artist - Title",
		},
		{
			Case:            "with all fields",
			ExpectedEnabled: true,
			ExpectedStatus:  "playing",
			JSONResponse: `{
				"title": "Title",
				"artist": "Artist",
				"artistUrl": "https://music.youtube.com/channel/test",
				"url": "https://music.youtube.com/watch?v=test",
				"mediaType": "AUDIO",
				"isPaused": false,
				"songDuration": 221,
				"elapsedSeconds": 100
			}`,
			ExpectedString: "Playing Artist - Title",
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)

		port := tc.Port
		if port == 0 {
			port = 26538
		}

		url := fmt.Sprintf("http://127.0.0.1:%d/api/v1/song-info", port)
		env.On("HTTPRequest", url).Return([]byte(tc.JSONResponse), tc.HTTPError)

		props := options.Map{
			PlayingIcon: "Playing ",
			PausedIcon:  "Paused ",
		}

		if tc.Port != 0 {
			props[Port] = tc.Port
		}

		peard := &PearDesktop{}
		peard.Init(props, env)

		enabled := peard.Enabled()
		assert.Equal(t, tc.ExpectedEnabled, enabled, tc.Case)

		if !tc.ExpectedEnabled {
			continue
		}

		assert.Equal(t, tc.ExpectedStatus, peard.Status, tc.Case)
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, peard.Template(), peard), tc.Case)
	}
}
