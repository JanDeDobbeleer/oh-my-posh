package segments

import (
	"fmt"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/cli/auth"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/template"

	"github.com/stretchr/testify/assert"
)

func TestPear(t *testing.T) {
	cases := []struct {
		HTTPError       error
		Case            string
		JSONResponse    string
		ExpectedString  string
		Port            int
		HasToken        bool
		ExpectedEnabled bool
	}{
		{
			Case:            "no token in cache",
			ExpectedEnabled: false,
		},
		{
			Case:            "no response",
			HasToken:        true,
			ExpectedEnabled: false,
			HTTPError:       assert.AnError,
		},
		{
			Case:            "empty response",
			HasToken:        true,
			ExpectedEnabled: false,
			JSONResponse:    "",
		},
		{
			Case:            "invalid response",
			HasToken:        true,
			ExpectedEnabled: false,
			JSONResponse:    "invalid json",
		},
		{
			Case:            "playing",
			HasToken:        true,
			ExpectedEnabled: true,
			JSONResponse:    `{"title": "Title", "artist": "Artist", "isPaused": false, "songDuration": 221, "elapsedSeconds": 100}`,
			ExpectedString:  "Playing Artist - Title",
		},
		{
			Case:            "paused",
			HasToken:        true,
			ExpectedEnabled: true,
			JSONResponse:    `{"title": "Title", "artist": "Artist", "isPaused": true, "songDuration": 221, "elapsedSeconds": 100}`,
			ExpectedString:  "Paused Artist - Title",
		},
		{
			Case:            "custom port",
			HasToken:        true,
			ExpectedEnabled: true,
			Port:            12345,
			JSONResponse:    `{"title": "Title", "artist": "Artist", "isPaused": false}`,
			ExpectedString:  "Playing Artist - Title",
		},
	}
	for _, tc := range cases {
		env := new(mock.Environment)

		port := tc.Port
		if port == 0 {
			port = defaultPearPort
		}

		url := fmt.Sprintf("http://127.0.0.1:%d/api/v1/song-info", port)
		env.On("HTTPRequest", url).Return([]byte(tc.JSONResponse), tc.HTTPError)

		if tc.HasToken {
			cache.Device.Set(auth.PEARTOKEN, "test_token", cache.INFINITE)
		}

		props := options.Map{
			PlayingIcon: template.RawMarkup("Playing "),
			PausedIcon:  template.RawMarkup("Paused "),
		}

		if tc.Port != 0 {
			props[PearPort] = tc.Port
		}

		pear := new(Pear)
		pear.Init(props, env)

		assert.Equal(t, tc.ExpectedEnabled, pear.Enabled(), tc.Case)
		cache.Device.DeleteAll()

		if !tc.ExpectedEnabled {
			continue
		}

		assert.Equal(t, tc.ExpectedString, renderTemplate(env, pear.Template(), pear), tc.Case)
	}
}
