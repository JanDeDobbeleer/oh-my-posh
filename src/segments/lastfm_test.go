package segments

import (
	"errors"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/stretchr/testify/assert"
)

const (
	LFMAPIURL = "https://ws.audioscrobbler.com/2.0/?method=user.getrecenttracks&api_key=key&user=KibbeWater&format=json&limit=1"
)

func TestLFMSegmentSingle(t *testing.T) {
	cases := []struct {
		Error           error
		Case            string
		APIJSONResponse string
		ExpectedString  string
		Template        string
		ExpectedEnabled bool
	}{
		{
			Case:            "All Defaults",
			APIJSONResponse: `{"recenttracks":{"track":[{"artist":{"#text":"C.Gambino"},"name":"Automatic","@attr":{"nowplaying":"true"}}]}}`,
			ExpectedString:  "\uE602 C.Gambino - Automatic",
			ExpectedEnabled: true,
		},
		{
			Case:            "Custom Template",
			APIJSONResponse: `{"recenttracks":{"track":[{"artist":{"#text":"C.Gambino"},"name":"Automatic","@attr":{"nowplaying":"true"}}]}}`,
			ExpectedString:  "\uE602 C.Gambino - Automatic",
			ExpectedEnabled: true,
			Template:        "{{ .Icon }}{{ if ne .Status \"stopped\" }}{{ .Full }}{{ end }}",
		},
		{
			Case:            "Song Stopped",
			APIJSONResponse: `{"recenttracks":{"track":[{"artist":{"#text":"C.Gambino"},"name":"Automatic","date":{"uts":"1699350223"}}]}}`,
			ExpectedString:  "\uF04D",
			ExpectedEnabled: true,
			Template:        "{{ .Icon }}",
		},
		{
			Case:            "Error in retrieving data",
			APIJSONResponse: "nonsense",
			Error:           errors.New("Something went wrong"),
			ExpectedEnabled: false,
		},
	}

	for _, tc := range cases {
		env := &mock.Environment{}
		props := properties.Map{
			APIKey:                 "key",
			Username:               "KibbeWater",
			properties.HTTPTimeout: 20000,
		}

		env.On("HTTPRequest", LFMAPIURL).Return([]byte(tc.APIJSONResponse), tc.Error)

		lfm := &LastFM{}
		lfm.Init(props, env)

		enabled := lfm.Enabled()
		assert.Equal(t, tc.ExpectedEnabled, enabled, tc.Case)
		if !enabled {
			continue
		}

		if tc.Template == "" {
			tc.Template = lfm.Template()
		}
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, tc.Template, lfm), tc.Case)
	}
}
