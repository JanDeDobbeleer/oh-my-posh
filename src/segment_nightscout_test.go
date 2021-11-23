package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	NSAPIURL = "FAKE"
)

func TestNSSegment(t *testing.T) {
	cases := []struct {
		Case            string
		JSONResponse    string
		ExpectedString  string
		ExpectedEnabled bool
		Template        string
		Error           error
	}{
		{
			Case: "Flat 150",
			JSONResponse: `
			[{"_id":"619d6fa819696e8ded5b2206","sgv":150,"date":1637707537000,"dateString":"2021-11-23T22:45:37.000Z","trend":4,"direction":"Flat","device":"share2","type":"sgv","utcOffset":0,"sysTime":"2021-11-23T22:45:37.000Z","mills":1637707537000}]`, // nolint:lll
			Template:        " {{.Sgv}}{{.TrendIcon}}",
			ExpectedString:  " 150→",
			ExpectedEnabled: true,
		},
		{
			Case: "DoubleDown 50",
			JSONResponse: `
			[{"_id":"619d6fa819696e8ded5b2206","sgv":50,"date":1637707537000,"dateString":"2021-11-23T22:45:37.000Z","trend":4,"direction":"DoubleDown","device":"share2","type":"sgv","utcOffset":0,"sysTime":"2021-11-23T22:45:37.000Z","mills":1637707537000}]`, // nolint:lll
			Template:        " {{.Sgv}}{{.TrendIcon}}",
			ExpectedString:  " 50↓↓",
			ExpectedEnabled: true,
		},
		{
			Case: "DoubleUp 250",
			JSONResponse: `
			[{"_id":"619d6fa819696e8ded5b2206","sgv":250,"date":1637707537000,"dateString":"2021-11-23T22:45:37.000Z","trend":4,"direction":"DoubleUp","device":"share2","type":"sgv","utcOffset":0,"sysTime":"2021-11-23T22:45:37.000Z","mills":1637707537000}]`, // nolint:lll
			Template:        " {{.Sgv}}{{.TrendIcon}}",
			ExpectedString:  " 250↑↑",
			ExpectedEnabled: true,
		},
		{
			Case: "SingleUp 130",
			JSONResponse: `
			[{"_id":"619d6fa819696e8ded5b2206","sgv":130,"date":1637707537000,"dateString":"2021-11-23T22:45:37.000Z","trend":4,"direction":"SingleUp","device":"share2","type":"sgv","utcOffset":0,"sysTime":"2021-11-23T22:45:37.000Z","mills":1637707537000}]`, // nolint:lll
			Template:        " {{.Sgv}}{{.TrendIcon}}",
			ExpectedString:  " 130↑",
			ExpectedEnabled: true,
		},
		{
			Case: "FortyFiveUp 174",
			JSONResponse: `
			[{"_id":"619d6fa819696e8ded5b2206","sgv":174,"date":1637707537000,"dateString":"2021-11-23T22:45:37.000Z","trend":4,"direction":"FortyFiveUp","device":"share2","type":"sgv","utcOffset":0,"sysTime":"2021-11-23T22:45:37.000Z","mills":1637707537000}]`, // nolint:lll
			Template:        " {{.Sgv}}{{.TrendIcon}}",
			ExpectedString:  " 174↗",
			ExpectedEnabled: true,
		},
		{
			Case: "FortyFiveDown 61",
			JSONResponse: `
			[{"_id":"619d6fa819696e8ded5b2206","sgv":61,"date":1637707537000,"dateString":"2021-11-23T22:45:37.000Z","trend":4,"direction":"FortyFiveDown","device":"share2","type":"sgv","utcOffset":0,"sysTime":"2021-11-23T22:45:37.000Z","mills":1637707537000}]`, // nolint:lll
			Template:        " {{.Sgv}}{{.TrendIcon}}",
			ExpectedString:  " 61↘",
			ExpectedEnabled: true,
		},
		{
			Case: "DoubleDown 50",
			JSONResponse: `
			[{"_id":"619d6fa819696e8ded5b2206","sgv":50,"date":1637707537000,"dateString":"2021-11-23T22:45:37.000Z","trend":4,"direction":"DoubleDown","device":"share2","type":"sgv","utcOffset":0,"sysTime":"2021-11-23T22:45:37.000Z","mills":1637707537000}]`, // nolint:lll
			Template:        " {{.Sgv}}{{.TrendIcon}}",
			ExpectedString:  " 50↓↓",
			ExpectedEnabled: true,
		},
		{
			Case:            "Error in retrieving data",
			JSONResponse:    "nonsense",
			Error:           errors.New("Something went wrong"),
			ExpectedEnabled: false,
		},
	}

	for _, tc := range cases {
		env := &MockedEnvironment{}
		props := &properties{
			values: map[Property]interface{}{
				CacheTimeout: 0,
				URL:          "FAKE",
			},
		}

		env.On("doGet", NSAPIURL).Return([]byte(tc.JSONResponse), tc.Error)

		if tc.Template != "" {
			props.values[SegmentTemplate] = tc.Template
		}

		ns := &nightscout{
			props: props,
			env:   env,
		}

		enabled := ns.enabled()
		assert.Equal(t, tc.ExpectedEnabled, enabled, tc.Case)
		if !enabled {
			continue
		}

		assert.Equal(t, tc.ExpectedString, ns.string(), tc.Case)
	}
}
