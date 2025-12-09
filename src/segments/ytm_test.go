package segments

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/cli/auth"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"

	"github.com/stretchr/testify/assert"
)

func TestYTM(t *testing.T) {
	cases := []struct {
		HTTPError       error
		Case            string
		JSONResponse    string
		ExpectedString  string
		HasToken        bool
		ExpectedEnabled bool
	}{
		{
			Case:            "no token in cache",
			HasToken:        false,
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
			Case:            "paused",
			HasToken:        true,
			ExpectedEnabled: true,
			JSONResponse:    `{"video": {"author": "Author", "title": "Title"}, "player": {"trackState": 0, "adPlaying": false}}`,
			ExpectedString:  "Paused Author - Title",
		},
		{
			Case:            "playing",
			HasToken:        true,
			ExpectedEnabled: true,
			JSONResponse:    `{"video": {"author": "Author", "title": "Title"}, "player": {"trackState": 1, "adPlaying": false}}`,
			ExpectedString:  "Playing Author - Title",
		},
		{
			Case:            "buffering",
			HasToken:        true,
			ExpectedEnabled: true,
			JSONResponse:    `{"video": {"author": "Author", "title": "Title"}, "player": {"trackState": 2, "adPlaying": false}}`,
			ExpectedString:  "Playing Author - Title",
		},
		{
			Case:            "stopped",
			HasToken:        true,
			ExpectedEnabled: true,
			JSONResponse:    `{"video": {"author": "Author", "title": "Title"}, "player": {"trackState": -1, "adPlaying": false}}`,
			ExpectedString:  "Stopped",
		},
		{
			Case:            "ad playing",
			HasToken:        true,
			ExpectedEnabled: true,
			JSONResponse:    `{"video": {"author": "Author", "title": "Title"}, "player": {"trackState": 1, "adPlaying": true}}`,
			ExpectedString:  "Ad Playing Author - Title",
		},
	}
	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("HTTPRequest", ytmdaStatusURL).Return([]byte(tc.JSONResponse), tc.HTTPError)

		if tc.HasToken {
			cache.Set(cache.Device, auth.YTMDATOKEN, "test_token", cache.INFINITE)
		}

		props := options.Map{
			StoppedIcon: "Stopped ",
			PlayingIcon: "Playing ",
			PausedIcon:  "Paused ",
			AdIcon:      "Ad ",
		}

		ytm := new(Ytm)
		ytm.Init(props, env)

		assert.Equal(t, tc.ExpectedEnabled, ytm.Enabled(), tc.Case)
		cache.DeleteAll(cache.Device)

		if !tc.ExpectedEnabled {
			continue
		}

		assert.Equal(t, tc.ExpectedString, renderTemplate(env, ytm.Template(), ytm), tc.Case)
	}
}
