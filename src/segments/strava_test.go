package segments

import (
	"errors"
	"testing"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/stretchr/testify/assert"
	testify_ "github.com/stretchr/testify/mock"
)

type mockedStravaAPI struct {
	testify_.Mock
}

func (s *mockedStravaAPI) GetActivities() ([]*StravaData, error) {
	args := s.Called()
	return args.Get(0).([]*StravaData), args.Error(1)
}

func TestStravaSegment(t *testing.T) {
	h, _ := time.ParseDuration("6h")
	sixHoursAgo := time.Now().Add(-h)
	h, _ = time.ParseDuration("100h")
	fourDaysAgo := time.Now().Add(-h)

	cases := []struct {
		APIError        error
		Case            string
		ExpectedString  string
		Template        string
		StravaData      []*StravaData
		ExpectedEnabled bool
	}{
		{
			Case: "Ride 6",
			StravaData: []*StravaData{
				{
					Type:      "Ride",
					StartDate: sixHoursAgo,
					Name:      "Sesongens første på tjukkas",
					Distance:  16144.0,
				},
			},
			Template:        "{{.Ago}} {{.Icon}}",
			ExpectedString:  "6h \uf206",
			ExpectedEnabled: true,
		},
		{
			Case: "Run 100",
			StravaData: []*StravaData{
				{
					Type:      "Run",
					StartDate: fourDaysAgo,
					Name:      "Sesongens første på tjukkas",
					Distance:  16144.0,
				},
			},
			Template:        "{{.Ago}} {{.Icon}}",
			ExpectedString:  "4d \ue213",
			ExpectedEnabled: true,
		},
		{
			Case:            "Error in retrieving data",
			APIError:        errors.New("Something went wrong"),
			ExpectedEnabled: false,
		},
		{
			Case:            "Empty array",
			StravaData:      []*StravaData{},
			ExpectedString:  noActivitiesFound,
			ExpectedEnabled: true,
		},
		{
			Case: "Faulty template",
			StravaData: []*StravaData{
				{
					Type:      "Run",
					StartDate: fourDaysAgo,
					Name:      "Sesongens første på tjukkas",
					Distance:  16144.0,
				},
			},
			Template:        "{{.Ago}}{{.Burp}}",
			ExpectedString:  "<.Data.Burp>: can't evaluate field Burp in type template.Data",
			ExpectedEnabled: true,
		},
	}

	for _, tc := range cases {
		api := &mockedStravaAPI{}
		api.On("GetActivities").Return(tc.StravaData, tc.APIError)

		strava := &Strava{
			api:   api,
			props: &properties.Map{},
		}

		enabled := strava.Enabled()
		assert.Equal(t, tc.ExpectedEnabled, enabled, tc.Case)
		if !enabled {
			continue
		}

		if tc.Template == "" {
			tc.Template = strava.Template()
		}

		var got = renderTemplate(&mock.Environment{}, tc.Template, strava)
		assert.Equal(t, tc.ExpectedString, got, tc.Case)
	}
}
