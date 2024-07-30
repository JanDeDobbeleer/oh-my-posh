package segments

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/stretchr/testify/assert"
	testify_ "github.com/stretchr/testify/mock"
)

const (
	CARBONINTENSITYURL = "https://api.carbonintensity.org.uk/intensity"
)

func TestCarbonIntensitySegmentSingle(t *testing.T) {
	cases := []struct {
		Case            string
		HasError        bool
		HasData         bool
		Actual          int
		Forecast        int
		Index           string
		ExpectedString  string
		ExpectedEnabled bool
		Template        string
	}{
		{
			Case:            "Very Low, Going Down",
			HasError:        false,
			HasData:         true,
			Actual:          20,
			Forecast:        10,
			Index:           "very low",
			ExpectedString:  "CO₂ ↓↓20 ↘ 10",
			ExpectedEnabled: true,
		},
		{
			Case:            "Very Low, Staying Same",
			HasError:        false,
			HasData:         true,
			Actual:          20,
			Forecast:        20,
			Index:           "very low",
			ExpectedString:  "CO₂ ↓↓20 → 20",
			ExpectedEnabled: true,
		},
		{
			Case:            "Very Low, Going Up",
			HasError:        false,
			HasData:         true,
			Actual:          20,
			Forecast:        30,
			Index:           "very low",
			ExpectedString:  "CO₂ ↓↓20 ↗ 30",
			ExpectedEnabled: true,
		},
		{
			Case:            "Low, Going Down",
			HasError:        false,
			HasData:         true,
			Actual:          100,
			Forecast:        50,
			Index:           "low",
			ExpectedString:  "CO₂ ↓100 ↘ 50",
			ExpectedEnabled: true,
		},
		{
			Case:            "Low, Staying Same",
			HasError:        false,
			HasData:         true,
			Actual:          100,
			Forecast:        100,
			Index:           "low",
			ExpectedString:  "CO₂ ↓100 → 100",
			ExpectedEnabled: true,
		},
		{
			Case:            "Low, Going Up",
			HasError:        false,
			HasData:         true,
			Actual:          100,
			Forecast:        150,
			Index:           "low",
			ExpectedString:  "CO₂ ↓100 ↗ 150",
			ExpectedEnabled: true,
		},
		{
			Case:            "Moderate, Going Down",
			HasError:        false,
			HasData:         true,
			Actual:          150,
			Forecast:        100,
			Index:           "moderate",
			ExpectedString:  "CO₂ •150 ↘ 100",
			ExpectedEnabled: true,
		},
		{
			Case:            "Moderate, Staying Same",
			HasError:        false,
			HasData:         true,
			Actual:          150,
			Forecast:        150,
			Index:           "moderate",
			ExpectedString:  "CO₂ •150 → 150",
			ExpectedEnabled: true,
		},
		{
			Case:            "Moderate, Going Up",
			HasError:        false,
			HasData:         true,
			Actual:          150,
			Forecast:        200,
			Index:           "moderate",
			ExpectedString:  "CO₂ •150 ↗ 200",
			ExpectedEnabled: true,
		},
		{
			Case:            "High, Going Down",
			HasError:        false,
			HasData:         true,
			Actual:          200,
			Forecast:        150,
			Index:           "high",
			ExpectedString:  "CO₂ ↑200 ↘ 150",
			ExpectedEnabled: true,
		},
		{
			Case:            "High, Staying Same",
			HasError:        false,
			HasData:         true,
			Actual:          200,
			Forecast:        200,
			Index:           "high",
			ExpectedString:  "CO₂ ↑200 → 200",
			ExpectedEnabled: true,
		},
		{
			Case:            "High, Going Up",
			HasError:        false,
			HasData:         true,
			Actual:          200,
			Forecast:        300,
			Index:           "high",
			ExpectedString:  "CO₂ ↑200 ↗ 300",
			ExpectedEnabled: true,
		},
		{
			Case:            "Missing Actual",
			HasError:        false,
			HasData:         true,
			Actual:          0, // Missing data will be parsed to the default value of 0
			Forecast:        300,
			Index:           "high",
			ExpectedString:  "CO₂ ↑?? → 300",
			ExpectedEnabled: true,
		},
		{
			Case:            "Missing Forecast",
			HasError:        false,
			HasData:         true,
			Actual:          200,
			Forecast:        0, // Missing data will be parsed to the default value of 0
			Index:           "high",
			ExpectedString:  "CO₂ ↑200 → ??",
			ExpectedEnabled: true,
		},
		{
			Case:            "Missing Index",
			HasError:        false,
			HasData:         true,
			Actual:          200,
			Forecast:        300,
			Index:           "", // Missing data will be parsed to the default value of ""
			ExpectedString:  "CO₂ 200 ↗ 300",
			ExpectedEnabled: true,
		},
		{
			Case:            "Missing Data",
			HasError:        false,
			HasData:         false,
			Actual:          0,
			Forecast:        0,
			Index:           "",
			ExpectedString:  "CO₂ ?? → ??",
			ExpectedEnabled: true,
		},
		{
			Case:            "Error",
			HasError:        true,
			HasData:         false,
			Actual:          0,
			Forecast:        0,
			Index:           "",
			ExpectedString:  "",
			ExpectedEnabled: false,
		},
	}

	for _, tc := range cases {
		env := &mock.Environment{}
		var props = properties.Map{
			properties.HTTPTimeout:  5000,
			properties.CacheTimeout: 0,
		}

		jsonResponse := fmt.Sprintf(
			`{ "data": [ { "from": "2023-10-27T12:30Z", "to": "2023-10-27T13:00Z", "intensity": { "forecast": %d, "actual": %d, "index": "%s" } } ] }`,
			tc.Forecast, tc.Actual, tc.Index,
		)

		if !tc.HasData {
			jsonResponse = `{ "data": [] }`
		}

		if tc.HasError {
			jsonResponse = `{ "error": "Something went wrong" }`
		}

		responseError := errors.New("Something went wrong")
		if !tc.HasError {
			responseError = nil
		}

		env.On("HTTPRequest", CARBONINTENSITYURL).Return([]byte(jsonResponse), responseError)
		env.On("Error", testify_.Anything)
		env.On("Flags").Return(&runtime.Flags{})

		d := &CarbonIntensity{
			props: props,
			env:   env,
		}

		enabled := d.Enabled()
		assert.Equal(t, tc.ExpectedEnabled, enabled, tc.Case)
		if !enabled {
			continue
		}

		if tc.Template == "" {
			tc.Template = d.Template()
		}
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, tc.Template, d), tc.Case)
	}
}
