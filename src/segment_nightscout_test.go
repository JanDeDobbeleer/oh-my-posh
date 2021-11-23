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
			[{"_id":"619d6fa819696e8ded5b2206","sgv":150,"date":1637707537000,"dateString":"2021-11-23T22:45:37.000Z","trend":4,"direction":"Flat","device":"share2","type":"sgv","utcOffset":0,"sysTime":"2021-11-23T22:45:37.000Z","mills":1637707537000}]`,
			Template:        " {{.Sgv}}{{.TrendIcon}}",
			ExpectedString:  " 150→",
			ExpectedEnabled: true,
		},
		{
			Case: "DoubleDown 50",
			JSONResponse: `
			[{"_id":"619d6fa819696e8ded5b2206","sgv":50,"date":1637707537000,"dateString":"2021-11-23T22:45:37.000Z","trend":4,"direction":"DoubleDown","device":"share2","type":"sgv","utcOffset":0,"sysTime":"2021-11-23T22:45:37.000Z","mills":1637707537000}]`,
			Template:        " {{.Sgv}}{{.TrendIcon}}",
			ExpectedString:  " 50↓↓",
			ExpectedEnabled: true,
		},
		{
			Case: "DoubleUp 250",
			JSONResponse: `
			[{"_id":"619d6fa819696e8ded5b2206","sgv":250,"date":1637707537000,"dateString":"2021-11-23T22:45:37.000Z","trend":4,"direction":"DoubleUp","device":"share2","type":"sgv","utcOffset":0,"sysTime":"2021-11-23T22:45:37.000Z","mills":1637707537000}]`,
			Template:        " {{.Sgv}}{{.TrendIcon}}",
			ExpectedString:  " 250↑↑",
			ExpectedEnabled: true,
		},
		{
			Case: "SingleUp 130",
			JSONResponse: `
			[{"_id":"619d6fa819696e8ded5b2206","sgv":130,"date":1637707537000,"dateString":"2021-11-23T22:45:37.000Z","trend":4,"direction":"SingleUp","device":"share2","type":"sgv","utcOffset":0,"sysTime":"2021-11-23T22:45:37.000Z","mills":1637707537000}]`,
			Template:        " {{.Sgv}}{{.TrendIcon}}",
			ExpectedString:  " 130↑",
			ExpectedEnabled: true,
		},
		{
			Case: "FortyFiveUp 174",
			JSONResponse: `
			[{"_id":"619d6fa819696e8ded5b2206","sgv":174,"date":1637707537000,"dateString":"2021-11-23T22:45:37.000Z","trend":4,"direction":"FortyFiveUp","device":"share2","type":"sgv","utcOffset":0,"sysTime":"2021-11-23T22:45:37.000Z","mills":1637707537000}]`,
			Template:        " {{.Sgv}}{{.TrendIcon}}",
			ExpectedString:  " 174↗",
			ExpectedEnabled: true,
		},
		{
			Case: "FortyFiveDown 61",
			JSONResponse: `
			[{"_id":"619d6fa819696e8ded5b2206","sgv":61,"date":1637707537000,"dateString":"2021-11-23T22:45:37.000Z","trend":4,"direction":"FortyFiveDown","device":"share2","type":"sgv","utcOffset":0,"sysTime":"2021-11-23T22:45:37.000Z","mills":1637707537000}]`,
			Template:        " {{.Sgv}}{{.TrendIcon}}",
			ExpectedString:  " 61↘",
			ExpectedEnabled: true,
		},
		{
			Case: "DoubleDown 50",
			JSONResponse: `
			[{"_id":"619d6fa819696e8ded5b2206","sgv":50,"date":1637707537000,"dateString":"2021-11-23T22:45:37.000Z","trend":4,"direction":"DoubleDown","device":"share2","type":"sgv","utcOffset":0,"sysTime":"2021-11-23T22:45:37.000Z","mills":1637707537000}]`,
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

/*
func TestOWMSegmentIcons(t *testing.T) {
	cases := []struct {
		Case               string
		IconID             string
		ExpectedIconString string
	}{
		{
			Case:               "Sunny Display day",
			IconID:             "01d",
			ExpectedIconString: "\ufa98",
		},
		{
			Case:               "Light clouds Display day",
			IconID:             "02d",
			ExpectedIconString: "\ufa94",
		},
		{
			Case:               "Cloudy Display day",
			IconID:             "03d",
			ExpectedIconString: "\ue33d",
		},
		{
			Case:               "Broken Clouds Display day",
			IconID:             "04d",
			ExpectedIconString: "\ue312",
		},
		{
			Case:               "Shower Rain Display day",
			IconID:             "09d",
			ExpectedIconString: "\ufa95",
		},
		{
			Case:               "Rain Display day",
			IconID:             "10d",
			ExpectedIconString: "\ue308",
		},
		{
			Case:               "Thunderstorm Display day",
			IconID:             "11d",
			ExpectedIconString: "\ue31d",
		},
		{
			Case:               "Snow Display day",
			IconID:             "13d",
			ExpectedIconString: "\ue31a",
		},
		{
			Case:               "Fog Display day",
			IconID:             "50d",
			ExpectedIconString: "\ue313",
		},

		{
			Case:               "Sunny Display night",
			IconID:             "01n",
			ExpectedIconString: "\ufa98",
		},
		{
			Case:               "Light clouds Display night",
			IconID:             "02n",
			ExpectedIconString: "\ufa94",
		},
		{
			Case:               "Cloudy Display night",
			IconID:             "03n",
			ExpectedIconString: "\ue33d",
		},
		{
			Case:               "Broken Clouds Display night",
			IconID:             "04n",
			ExpectedIconString: "\ue312",
		},
		{
			Case:               "Shower Rain Display night",
			IconID:             "09n",
			ExpectedIconString: "\ufa95",
		},
		{
			Case:               "Rain Display night",
			IconID:             "10n",
			ExpectedIconString: "\ue308",
		},
		{
			Case:               "Thunderstorm Display night",
			IconID:             "11n",
			ExpectedIconString: "\ue31d",
		},
		{
			Case:               "Snow Display night",
			IconID:             "13n",
			ExpectedIconString: "\ue31a",
		},
		{
			Case:               "Fog Display night",
			IconID:             "50n",
			ExpectedIconString: "\ue313",
		},
	}

	for _, tc := range cases {
		env := &MockedEnvironment{}
		props := &properties{
			values: map[Property]interface{}{
				APIKey:       "key",
				Location:     "AMSTERDAM,NL",
				Units:        "metric",
				CacheTimeout: 0,
			},
		}

		response := fmt.Sprintf(`{"weather":[{"icon":"%s"}],"main":{"temp":20}}`, tc.IconID)
		expectedString := fmt.Sprintf("%s (20°C)", tc.ExpectedIconString)

		env.On("doGet", OWMAPIURL).Return([]byte(response), nil)

		o := &owm{
			props: props,
			env:   env,
		}

		assert.Nil(t, o.setStatus())
		assert.Equal(t, expectedString, o.string(), tc.Case)
	}

	// test with hyperlink enabled
	for _, tc := range cases {
		env := &MockedEnvironment{}
		props := &properties{
			values: map[Property]interface{}{
				APIKey:          "key",
				Location:        "AMSTERDAM,NL",
				Units:           "metric",
				CacheTimeout:    0,
				EnableHyperlink: true,
			},
		}

		response := fmt.Sprintf(`{"weather":[{"icon":"%s"}],"main":{"temp":20}}`, tc.IconID)
		expectedString := fmt.Sprintf("[%s (20°C)](http://api.openweathermap.org/data/2.5/weather?q=AMSTERDAM,NL&units=metric&appid=key)", tc.ExpectedIconString)

		env.On("doGet", OWMAPIURL).Return([]byte(response), nil)

		props.values[SegmentTemplate] = "[{{.Weather}} ({{.Temperature}}{{.UnitIcon}})]({{.URL}})"

		o := &owm{
			props: props,
			env:   env,
		}

		assert.Nil(t, o.setStatus())
		assert.Equal(t, expectedString, o.string(), tc.Case)
	}
}
*/
