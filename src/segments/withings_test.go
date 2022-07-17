package segments

import (
	"errors"
	"oh-my-posh/mock"
	"oh-my-posh/properties"
	"testing"

	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
)

type mockedWithingsAPI struct {
	mock2.Mock
}

func (s *mockedWithingsAPI) GetMeasures(meastypes string) (*WithingsData, error) {
	args := s.Called(meastypes)
	return args.Get(0).(*WithingsData), args.Error(1)
}

func (s *mockedWithingsAPI) GetActivities(activities string) (*WithingsData, error) {
	args := s.Called(activities)
	return args.Get(0).(*WithingsData), args.Error(1)
}

func (s *mockedWithingsAPI) GetSleep() (*WithingsData, error) {
	args := s.Called()
	return args.Get(0).(*WithingsData), args.Error(1)
}

func TestWithingsSegment(t *testing.T) {
	cases := []struct {
		Case            string
		ExpectedString  string
		ExpectedEnabled bool
		Template        string
		MeasuresError   error
		ActivitiesError error
		SleepError      error
		WithingsData    *WithingsData
	}{
		{
			Case:            "Error",
			MeasuresError:   errors.New("error"),
			ActivitiesError: errors.New("error"),
			SleepError:      errors.New("error"),
			ExpectedEnabled: false,
		},
		{
			Case: "Only Measures data",
			WithingsData: &WithingsData{
				Body: &Body{
					MeasureGroups: []*MeasureGroup{
						{
							Measures: []*Measure{
								{
									Value: 7077,
									Unit:  -2,
								},
							},
						},
					},
				},
			},
			ActivitiesError: errors.New("error"),
			SleepError:      errors.New("error"),
			ExpectedEnabled: true,
			ExpectedString:  "70.77kg",
		},
		{
			Case: "Measures, no data",
			WithingsData: &WithingsData{
				Body: &Body{},
			},
			ActivitiesError: errors.New("error"),
			SleepError:      errors.New("error"),
			ExpectedEnabled: false,
		},
		{
			Case:           "Activities",
			Template:       "{{ .Steps }} steps",
			ExpectedString: "7077 steps",
			WithingsData: &WithingsData{
				Body: &Body{
					Activities: []*Activity{
						{
							Steps: 7077,
						},
					},
				},
			},
			MeasuresError:   errors.New("error"),
			SleepError:      errors.New("error"),
			ExpectedEnabled: true,
		},
		{
			Case:           "Sleep",
			Template:       "{{ .SleepHours }}hr",
			ExpectedString: "11.8hr",
			WithingsData: &WithingsData{
				Body: &Body{
					Series: []*Series{
						{
							Startdate: 1594159200,
							Enddate:   1594201500,
						},
					},
				},
			},
			MeasuresError:   errors.New("error"),
			ActivitiesError: errors.New("error"),
			ExpectedEnabled: true,
		},
		{
			Case:           "Sleep and Activity",
			Template:       "{{ .Steps }} steps with {{ .SleepHours }}hr of sleep",
			ExpectedString: "976 steps with 11.8hr of sleep",
			WithingsData: &WithingsData{
				Body: &Body{
					Series: []*Series{
						{
							Startdate: 1594159200,
							Enddate:   1594201500,
						},
					},
					Activities: []*Activity{
						{
							Steps: 976,
						},
					},
				},
			},
			MeasuresError:   errors.New("error"),
			ExpectedEnabled: true,
		},
	}

	for _, tc := range cases {
		api := &mockedWithingsAPI{}
		api.On("GetMeasures", "1").Return(tc.WithingsData, tc.MeasuresError)
		api.On("GetActivities", "steps").Return(tc.WithingsData, tc.ActivitiesError)
		api.On("GetSleep").Return(tc.WithingsData, tc.SleepError)

		withings := &Withings{
			api:   api,
			props: &properties.Map{},
		}

		enabled := withings.Enabled()
		assert.Equal(t, tc.ExpectedEnabled, enabled, tc.Case)
		if !enabled {
			continue
		}

		if tc.Template == "" {
			tc.Template = withings.Template()
		}

		var got = renderTemplate(&mock.MockedEnvironment{}, tc.Template, withings)
		assert.Equal(t, tc.ExpectedString, got, tc.Case)
	}
}
