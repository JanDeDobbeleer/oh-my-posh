package segments

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/template"

	"github.com/stretchr/testify/assert"
)

const (
	FAKEAPIURL = "FAKE"
)

func TestNSSegment(t *testing.T) {
	cases := []struct {
		Error           error
		Case            string
		JSONResponse    string
		ExpectedString  string
		Template        string
		CacheTimeout    int
		ExpectedEnabled bool
		CacheFoundFail  bool
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
			Case: "Float date value",
			JSONResponse: `
			[{"_id":"619d6fa819696e8ded5b2206","sgv":124,"date":1770512410938.386,"dateString":"2026-02-08T01:00:10.000Z","trend":4,"direction":"Flat","device":"share2","type":"sgv","utcOffset":0,"sysTime":"2026-02-08T01:00:10.000Z","mills":1770512410000}]`, //nolint:lll
			Template:        "\ue2a1 {{.Sgv}}{{.TrendIcon}}",
			ExpectedString:  "\ue2a1 124→",
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
			Case: "Error parsing response",
			JSONResponse: `
			4tffgt4e4567`,
			Template:        "\ue2a1 {{.Sgv}}{{.TrendIcon}}",
			ExpectedString:  "\ue2a1 50↓↓",
			ExpectedEnabled: false,
		},
		{
			Case: "Faulty template",
			JSONResponse: `
			[{"sgv":50,"direction":"DoubleDown"}]`,
			Template:        "\ue2a1 {{.Sgv}}{{.Burp}}",
			ExpectedString:  template.IncorrectTemplate,
			ExpectedEnabled: true,
		},
	}

	for _, tc := range cases {
		env := &mock.Environment{}
		props := options.Map{
			URL:     "FAKE",
			Headers: map[string]string{"Fake-Header": "xxxxx"},
		}

		env.On("HTTPRequest", FAKEAPIURL).Return([]byte(tc.JSONResponse), tc.Error)

		ns := &Nightscout{}
		ns.Init(props, env)

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

func TestNightscoutDataUnmarshalJSON(t *testing.T) {
	cases := []struct {
		Case         string
		JSONInput    string
		ExpectedDate int64
		ExpectError  bool
	}{
		{
			Case:         "Integer date value",
			JSONInput:    `{"date": 1637707537000}`,
			ExpectedDate: 1637707537000,
		},
		{
			Case:         "Floating-point date value",
			JSONInput:    `{"date": 1637707537000.5}`,
			ExpectedDate: 1637707537000,
		},
		{
			Case:         "Floating-point date with larger decimal",
			JSONInput:    `{"date": 1637707537123.789}`,
			ExpectedDate: 1637707537123,
		},
		{
			Case:        "Invalid date value",
			JSONInput:   `{"date": "not-a-number"}`,
			ExpectError: true,
		},
		{
			Case:      "Missing date field",
			JSONInput: `{"sgv": 150}`,
		},
	}

	for _, tc := range cases {
		var data NightscoutData
		err := json.Unmarshal([]byte(tc.JSONInput), &data)

		if tc.ExpectError {
			assert.Error(t, err, tc.Case)
			continue
		}

		assert.NoError(t, err, tc.Case)
		assert.Equal(t, tc.ExpectedDate, data.Date, tc.Case)
	}
}
