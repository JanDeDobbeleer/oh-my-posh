package segments

import (
	"errors"
	"fmt"
	"net/url"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/stretchr/testify/assert"
	testify_ "github.com/stretchr/testify/mock"
)

const (
	OWMWEATHERAPIURL = "https://api.openweathermap.org/data/2.5/weather?q=%s&units=metric&appid=key"
)

func TestOWMSegmentSingle(t *testing.T) {
	cases := []struct {
		Error               error
		Case                string
		Location            string
		WeatherJSONResponse string
		ExpectedString      string
		Template            string
		ExpectedEnabled     bool
	}{
		{
			Case:                "Sunny Display",
			Location:            "AMSTERDAM,NL",
			WeatherJSONResponse: `{"weather":[{"icon":"01d"}],"main":{"temp":20}}`,
			ExpectedString:      "\ue30d (20°C)",
			ExpectedEnabled:     true,
		},
		{
			Case:                "Sunny Display",
			Location:            "AMSTERDAM,NL",
			WeatherJSONResponse: `{"weather":[{"icon":"01d"}],"main":{"temp":20}}`,
			ExpectedString:      "\ue30d (20°C)",
			ExpectedEnabled:     true,
			Template:            "{{.Weather}} ({{.Temperature}}{{.UnitIcon}})",
		},
		{
			Case:                "Sunny Display",
			Location:            "AMSTERDAM,NL",
			WeatherJSONResponse: `{"weather":[{"icon":"01d"}],"main":{"temp":20}}`,
			ExpectedString:      "\ue30d",
			ExpectedEnabled:     true,
			Template:            "{{.Weather}} ",
		},
		{
			Case:                "Config Skip Geocoding Check With Location",
			Location:            "AMSTERDAM,NL",
			WeatherJSONResponse: `{"weather":[{"icon":"01d"}],"main":{"temp":20}}`,
			ExpectedString:      "\ue30d (20°C)",
			ExpectedEnabled:     true,
		},
		{
			Case:                "Config Skip Geocoding Check Without Location",
			WeatherJSONResponse: `{"weather":[{"icon":"01d"}],"main":{"temp":20}}`,
			ExpectedEnabled:     false,
		},
		{
			Case:                "Error in retrieving data",
			Location:            "AMSTERDAM,NL",
			WeatherJSONResponse: "nonsense",
			Error:               errors.New("Something went wrong"),
			ExpectedEnabled:     false,
		},
	}

	for _, tc := range cases {
		env := &mock.Environment{}
		props := properties.Map{
			APIKey:                  "key",
			Location:                tc.Location,
			Units:                   "metric",
			properties.CacheTimeout: 0,
		}

		location := url.QueryEscape(tc.Location)
		testURL := fmt.Sprintf(OWMWEATHERAPIURL, location)
		env.On("HTTPRequest", testURL).Return([]byte(tc.WeatherJSONResponse), tc.Error)
		env.On("Error", testify_.Anything)

		o := &Owm{
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

func TestOWMSegmentIcons(t *testing.T) {
	cases := []struct {
		Case               string
		IconID             string
		ExpectedIconString string
	}{
		{
			Case:               "Sunny Display day",
			IconID:             "01d",
			ExpectedIconString: "\ue30d",
		},
		{
			Case:               "Light clouds Display day",
			IconID:             "02d",
			ExpectedIconString: "\ue302",
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
			ExpectedIconString: "\ue319",
		},
		{
			Case:               "Rain Display day",
			IconID:             "10d",
			ExpectedIconString: "\ue308",
		},
		{
			Case:               "Thunderstorm Display day",
			IconID:             "11d",
			ExpectedIconString: "\ue30f",
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
			ExpectedIconString: "\ue32b",
		},
		{
			Case:               "Light clouds Display night",
			IconID:             "02n",
			ExpectedIconString: "\ue37e",
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
			ExpectedIconString: "\ue319",
		},
		{
			Case:               "Rain Display night",
			IconID:             "10n",
			ExpectedIconString: "\ue325",
		},
		{
			Case:               "Thunderstorm Display night",
			IconID:             "11n",
			ExpectedIconString: "\ue32a",
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

	location := url.QueryEscape("AMSTERDAM,NL")
	testURL := fmt.Sprintf(OWMWEATHERAPIURL, location)

	for _, tc := range cases {
		env := &mock.Environment{}

		weatherResponse := fmt.Sprintf(`{"weather":[{"icon":"%s"}],"main":{"temp":20.3}}`, tc.IconID)
		expectedString := fmt.Sprintf("%s (20°C)", tc.ExpectedIconString)

		env.On("HTTPRequest", testURL).Return([]byte(weatherResponse), nil)

		o := &Owm{
			props: properties.Map{
				APIKey:                  "key",
				Location:                "AMSTERDAM,NL",
				Units:                   "metric",
				properties.CacheTimeout: 0,
			},
			env: env,
		}

		assert.Nil(t, o.setStatus())
		assert.Equal(t, expectedString, renderTemplate(env, o.Template(), o), tc.Case)
	}

	// test with hyperlink enabled
	for _, tc := range cases {
		env := &mock.Environment{}

		weatherResponse := fmt.Sprintf(`{"weather":[{"icon":"%s"}],"main":{"temp":20.3}}`, tc.IconID)
		expectedString := fmt.Sprintf("«%s (20°C)»(%s)", tc.ExpectedIconString, testURL)

		env.On("HTTPRequest", testURL).Return([]byte(weatherResponse), nil)

		o := &Owm{
			props: properties.Map{
				APIKey:                  "key",
				Location:                "AMSTERDAM,NL",
				Units:                   "metric",
				properties.CacheTimeout: 0,
			},
			env: env,
		}

		assert.Nil(t, o.setStatus())
		assert.Equal(t, expectedString, renderTemplate(env, "«{{.Weather}} ({{.Temperature}}{{.UnitIcon}})»({{.URL}})", o), tc.Case)
	}
}
