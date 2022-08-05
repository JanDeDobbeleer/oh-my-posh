package segments

import (
	"errors"
	"oh-my-posh/mock"
	"oh-my-posh/properties"
	"oh-my-posh/template"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	FAKEAPIURL = "FAKE"
)

func TestNSSegment(t *testing.T) {
	cases := []struct {
		Case            string
		JSONResponse    string
		ExpectedString  string
		ExpectedEnabled bool
		CacheTimeout    int
		CacheFoundFail  bool
		Template        string
		Error           error
	}{
		{
			Case: "Flat 150",
			JSONResponse: `
			[{"_id":"619d6fa819696e8ded5b2206","sgv":150,"date":1637707537000,"dateString":"2021-11-23T22:45:37.000Z","trend":4,"direction":"Flat","device":"share2","type":"sgv","utcOffset":0,"sysTime":"2021-11-23T22:45:37.000Z","mills":1637707537000}]`, //nolint:lll
			Template:        "\ue2a1 {{.Sgv}}{{.TrendIcon}}",
			ExpectedString:  "\ue2a1 150→",
			ExpectedEnabled: true,
		},
		{
			Case: "DoubleDown 50",
			JSONResponse: `
			[{"_id":"619d6fa819696e8ded5b2206","sgv":50,"date":1637707537000,"dateString":"2021-11-23T22:45:37.000Z","trend":4,"direction":"DoubleDown","device":"share2","type":"sgv","utcOffset":0,"sysTime":"2021-11-23T22:45:37.000Z","mills":1637707537000}]`, //nolint:lll
			Template:        "\ue2a1 {{.Sgv}}{{.TrendIcon}}",
			ExpectedString:  "\ue2a1 50↓↓",
			ExpectedEnabled: true,
		},
		{
			Case: "DoubleUp 250",
			JSONResponse: `
			[{"_id":"619d6fa819696e8ded5b2206","sgv":250,"date":1637707537000,"dateString":"2021-11-23T22:45:37.000Z","trend":4,"direction":"DoubleUp","device":"share2","type":"sgv","utcOffset":0,"sysTime":"2021-11-23T22:45:37.000Z","mills":1637707537000}]`, //nolint:lll
			Template:        "\ue2a1 {{.Sgv}}{{.TrendIcon}}",
			ExpectedString:  "\ue2a1 250↑↑",
			ExpectedEnabled: true,
		},
		{
			Case: "SingleUp 130",
			JSONResponse: `
			[{"_id":"619d6fa819696e8ded5b2206","sgv":130,"date":1637707537000,"dateString":"2021-11-23T22:45:37.000Z","trend":4,"direction":"SingleUp","device":"share2","type":"sgv","utcOffset":0,"sysTime":"2021-11-23T22:45:37.000Z","mills":1637707537000}]`, //nolint:lll
			Template:        "\ue2a1 {{.Sgv}}{{.TrendIcon}}",
			ExpectedString:  "\ue2a1 130↑",
			ExpectedEnabled: true,
		},
		{
			Case: "FortyFiveUp 174",
			JSONResponse: `
			[{"_id":"619d6fa819696e8ded5b2206","sgv":174,"date":1637707537000,"dateString":"2021-11-23T22:45:37.000Z","trend":4,"direction":"FortyFiveUp","device":"share2","type":"sgv","utcOffset":0,"sysTime":"2021-11-23T22:45:37.000Z","mills":1637707537000}]`, //nolint:lll
			Template:        "\ue2a1 {{.Sgv}}{{.TrendIcon}}",
			ExpectedString:  "\ue2a1 174↗",
			ExpectedEnabled: true,
		},
		{
			Case: "FortyFiveDown 61",
			JSONResponse: `
			[{"_id":"619d6fa819696e8ded5b2206","sgv":61,"date":1637707537000,"dateString":"2021-11-23T22:45:37.000Z","trend":4,"direction":"FortyFiveDown","device":"share2","type":"sgv","utcOffset":0,"sysTime":"2021-11-23T22:45:37.000Z","mills":1637707537000}]`, //nolint:lll
			Template:        "\ue2a1 {{.Sgv}}{{.TrendIcon}}",
			ExpectedString:  "\ue2a1 61↘",
			ExpectedEnabled: true,
		},
		{
			Case: "DoubleDown 50",
			JSONResponse: `
			[{"_id":"619d6fa819696e8ded5b2206","sgv":50,"date":1637707537000,"dateString":"2021-11-23T22:45:37.000Z","trend":4,"direction":"DoubleDown","device":"share2","type":"sgv","utcOffset":0,"sysTime":"2021-11-23T22:45:37.000Z","mills":1637707537000}]`, //nolint:lll
			Template:        "\ue2a1 {{.Sgv}}{{.TrendIcon}}",
			ExpectedString:  "\ue2a1 50↓↓",
			ExpectedEnabled: true,
		},
		{
			Case:            "Error in retrieving data",
			JSONResponse:    "nonsense",
			Error:           errors.New("Something went wrong"),
			ExpectedEnabled: false,
		},
		{
			Case:            "Empty array",
			JSONResponse:    "[]",
			ExpectedEnabled: false,
		},
		{
			Case: "DoubleDown 50 from cache",
			JSONResponse: `
			[{"_id":"619d6fa819696e8ded5b2206","sgv":50,"date":1637707537000,"dateString":"2021-11-23T22:45:37.000Z","trend":4,"direction":"DoubleDown","device":"share2","type":"sgv","utcOffset":0,"sysTime":"2021-11-23T22:45:37.000Z","mills":1637707537000}]`, //nolint:lll
			Template:        "\ue2a1 {{.Sgv}}{{.TrendIcon}}",
			ExpectedString:  "\ue2a1 50↓↓",
			ExpectedEnabled: true,
			CacheTimeout:    10,
		},
		{
			Case: "DoubleDown 50 from cache not found",
			JSONResponse: `
			[{"_id":"619d6fa819696e8ded5b2206","sgv":50,"date":1637707537000,"dateString":"2021-11-23T22:45:37.000Z","trend":4,"direction":"DoubleDown","device":"share2","type":"sgv","utcOffset":0,"sysTime":"2021-11-23T22:45:37.000Z","mills":1637707537000}]`, //nolint:lll
			Template:        "\ue2a1 {{.Sgv}}{{.TrendIcon}}",
			ExpectedString:  "\ue2a1 50↓↓",
			ExpectedEnabled: true,
			CacheTimeout:    10,
			CacheFoundFail:  true,
		},
		{
			Case: "Error parsing response",
			JSONResponse: `
			4tffgt4e4567`,
			Template:        "\ue2a1 {{.Sgv}}{{.TrendIcon}}",
			ExpectedString:  "\ue2a1 50↓↓",
			ExpectedEnabled: false,
			CacheTimeout:    10,
		},
		{
			Case: "Faulty template",
			JSONResponse: `
			[{"sgv":50,"direction":"DoubleDown"}]`,
			Template:        "\ue2a1 {{.Sgv}}{{.Burp}}",
			ExpectedString:  template.IncorrectTemplate,
			ExpectedEnabled: true,
			CacheTimeout:    10,
		},
	}

	for _, tc := range cases {
		env := &mock.MockedEnvironment{}
		props := properties.Map{
			properties.CacheTimeout: tc.CacheTimeout,
			URL:                     "FAKE",
		}

		cache := &mock.MockedCache{}
		cache.On("Get", FAKEAPIURL).Return(tc.JSONResponse, !tc.CacheFoundFail)
		cache.On("Set", FAKEAPIURL, tc.JSONResponse, tc.CacheTimeout).Return()

		env.On("HTTPRequest", FAKEAPIURL).Return([]byte(tc.JSONResponse), tc.Error)
		env.On("Cache").Return(cache)

		ns := &Nightscout{
			props: props,
			env:   env,
		}

		enabled := ns.Enabled()
		assert.Equal(t, tc.ExpectedEnabled, enabled, tc.Case)
		if !enabled {
			continue
		}

		if tc.Template == "" {
			tc.Template = ns.Template()
		}
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, tc.Template, ns), tc.Case)
	}
}
