package main

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOWMSegmentSingle(t *testing.T) {
	cases := []struct {
		Case            string
		JSONResponse    string
		ExpectedString  string
		ExpectedEnabled bool
		Error           error
	}{
		{
			Case:            "Sunny Display",
			JSONResponse:    `{"weather":[{"icon":"01d"}],"main":{"temp":20}}`,
			ExpectedString:  "\ufa98 (20°C)",
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
				APIKEY:   "key",
				LOCATION: "AMSTERDAM,NL",
				UNITS:    "metric",
			},
		}

		url := "http://api.openweathermap.org/data/2.5/weather?q=AMSTERDAM,NL&units=metric&appid=key"

		env.On("doGet", url).Return([]byte(tc.JSONResponse), tc.Error)

		o := &owm{
			props: props,
			env:   env,
		}

		enabled := o.enabled()
		assert.Equal(t, tc.ExpectedEnabled, enabled, tc.Case)
		if !enabled {
			continue
		}

		assert.Equal(t, tc.ExpectedString, o.string(), tc.Case)
	}
}

func TestOWMSegmentIcons(t *testing.T) {
	cases := []struct {
		Case               string
		IconID             string
		ExpectedIconString string
	}{
		{
			Case:               "Sunny Display",
			IconID:             "01d",
			ExpectedIconString: "\ufa98",
		},
		{
			Case:               "Light clouds Display",
			IconID:             "02d",
			ExpectedIconString: "\ufa94",
		},
		{
			Case:               "Cloudy Display",
			IconID:             "03d",
			ExpectedIconString: "\ue33d",
		},
		{
			Case:               "Broken Clouds Display",
			IconID:             "04d",
			ExpectedIconString: "\ue312",
		},
		{
			Case:               "Shower Rain Display",
			IconID:             "09d",
			ExpectedIconString: "\ufa95",
		},
		{
			Case:               "Rain Display",
			IconID:             "10d",
			ExpectedIconString: "\ue308",
		},
		{
			Case:               "Thunderstorm Display",
			IconID:             "11d",
			ExpectedIconString: "\ue31d",
		},
		{
			Case:               "Snow Display",
			IconID:             "13d",
			ExpectedIconString: "\ue31a",
		},
		{
			Case:               "Fog Display",
			IconID:             "50d",
			ExpectedIconString: "\ue313",
		},
	}

	for _, tc := range cases {
		env := &MockedEnvironment{}
		props := &properties{
			values: map[Property]interface{}{
				APIKEY:   "key",
				LOCATION: "AMSTERDAM,NL",
				UNITS:    "metric",
			},
		}

		url := "http://api.openweathermap.org/data/2.5/weather?q=AMSTERDAM,NL&units=metric&appid=key"
		response := fmt.Sprintf(`{"weather":[{"icon":"%s"}],"main":{"temp":20}}`, tc.IconID)
		expectedString := fmt.Sprintf("%s (20°C)", tc.ExpectedIconString)

		env.On("doGet", url).Return([]byte(response), nil)

		o := &owm{
			props: props,
			env:   env,
		}

		assert.Nil(t, o.setStatus())
		assert.Equal(t, expectedString, o.string(), tc.Case)
	}
}
