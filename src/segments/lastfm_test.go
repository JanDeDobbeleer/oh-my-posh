package segments

import (
	"errors"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/stretchr/testify/assert"
	testify_ "github.com/stretchr/testify/mock"
)

const (
	LFMAPIURL = "https://ws.audioscrobbler.com/2.0/?method=user.getrecenttracks&api_key=key&user=KibbeWater&format=json&limit=1"
)

func TestLFMSegmentSingle(t *testing.T) {
	cases := []struct {
		Case            string
		APIJSONResponse string
		ExpectedString  string
		ExpectedEnabled bool
		Template        string
		Error           error
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
		var props properties.Map = properties.Map{
			APIKey:                  "key",
			Username:                "KibbeWater",
			properties.CacheTimeout: 0,
			properties.HTTPTimeout:  20000,
		}

		env.On("HTTPRequest", LFMAPIURL).Return([]byte(tc.APIJSONResponse), tc.Error)
		env.On("Error", testify_.Anything)

		o := &LastFM{
			props: props,
			env:   env,
		}

		enabled := o.Enabled()
		assert.Equal(t, tc.ExpectedEnabled, enabled, tc.Case)
		if !enabled {
			continue
		}

		if tc.Template == "" {
			tc.Template = o.Template()
		}
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, tc.Template, o), tc.Case)
	}
}
