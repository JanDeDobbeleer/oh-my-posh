package main

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	OWMAPIURL = "http://api.openweathermap.org/data/2.5/weather?q=AMSTERDAM,NL&units=metric&appid=key"
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
				APIKey:       "key",
				Location:     "AMSTERDAM,NL",
				Units:        "metric",
				CacheTimeout: 0,
			},
		}

		env.On("doGet", OWMAPIURL).Return([]byte(tc.JSONResponse), tc.Error)

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
}
func TestOWMSegmentFromCache(t *testing.T) {
	response := fmt.Sprintf(`{"weather":[{"icon":"%s"}],"main":{"temp":20}}`, "01d")
	expectedString := fmt.Sprintf("%s (20°C)", "\ufa98")

	env := &MockedEnvironment{}
	cache := &MockedCache{}
	props := &properties{
		values: map[Property]interface{}{
			APIKey:   "key",
			Location: "AMSTERDAM,NL",
			Units:    "metric",
		},
	}
	o := &owm{
		props: props,
		env:   env,
	}
	cache.On("get", "owm_response").Return(response, true)
	cache.On("set").Return()
	env.On("cache", nil).Return(cache)
	// env.On("doGet", "http://api.openweathermap.org/data/2.5/weather?q=AMSTERDAM,NL&units=metric&appid=key").Return([]byte(response), nil)

	assert.Nil(t, o.setStatus())
	assert.Equal(t, expectedString, o.string())
}
