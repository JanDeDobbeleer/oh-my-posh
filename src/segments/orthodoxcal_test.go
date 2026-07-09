package segments

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"

	"github.com/stretchr/testify/assert"
)

func TestOrthodoxCalEnabled(t *testing.T) {
	cases := []struct {
		Case                  string
		JSONResponse          string
		ExpectedFastLevelDesc string
		ExpectedSummaryTitle  string
		ExpectedEnabled       bool
	}{
		{
			Case: "no fast day",
			JSONResponse: `{
				"fast_level": 0,
				"fast_level_desc": "No Fast",
				"fast_exception": 0,
				"fast_exception_desc": "",
				"summary_title": "Hieromartyr Pancratius",
				"feast_level": 0,
				"feast_level_description": "Liturgy",
				"feasts": null,
				"saints": ["Hieromartyr Pancratius"],
				"titles": ["Thursday of the 6th week after Pentecost"],
				"tone": 4
			}`,
			ExpectedEnabled:       true,
			ExpectedFastLevelDesc: "No Fast",
			ExpectedSummaryTitle:  "Hieromartyr Pancratius",
		},
		{
			Case: "strict fast day",
			JSONResponse: `{
				"fast_level": 1,
				"fast_level_desc": "Strict Fast (Dry Fast)",
				"fast_exception": 2,
				"fast_exception_desc": "Wine and Oil are Allowed",
				"summary_title": "Dormition Fast",
				"feast_level": 0,
				"feast_level_description": "Liturgy",
				"feasts": null,
				"saints": ["Prophet Micah"],
				"titles": ["Monday of the 8th week after Pentecost"],
				"tone": 6
			}`,
			ExpectedEnabled:       true,
			ExpectedFastLevelDesc: "Strict Fast (Dry Fast)",
			ExpectedSummaryTitle:  "Dormition Fast",
		},
		{
			Case:            "http error",
			JSONResponse:    "",
			ExpectedEnabled: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Case, func(t *testing.T) {
			env := new(mock.Environment)

			opts := options.Map{
				OrthodoxCalType: "gregorian",
			}

			if tc.JSONResponse != "" {
				env.On("HTTPRequest", "https://orthocal.info/api/gregorian/").Return([]byte(tc.JSONResponse), nil)
			} else {
				env.On("HTTPRequest", "https://orthocal.info/api/gregorian/").Return([]byte{}, assert.AnError)
			}

			o := &OrthodoxCal{}
			o.Init(opts, env)

			enabled := o.Enabled()
			assert.Equal(t, tc.ExpectedEnabled, enabled, tc.Case)

			if enabled {
				assert.Equal(t, tc.ExpectedFastLevelDesc, o.FastLevelDesc, tc.Case)
				assert.Equal(t, tc.ExpectedSummaryTitle, o.SummaryTitle, tc.Case)
			}
		})
	}
}

func TestOrthodoxCalIsFasting(t *testing.T) {
	o := new(OrthodoxCal)
	assert.False(t, o.IsFasting())

	o.FastLevel = 3
	assert.True(t, o.IsFasting())
}

func TestOrthodoxCalFeastNames(t *testing.T) {
	o := new(OrthodoxCal)
	o.Feasts = []string{"Nativity of Christ", "Synaxis of the Theotokos"}
	assert.Equal(t, "Nativity of Christ, Synaxis of the Theotokos", o.FeastNames())

	o.Feasts = nil
	assert.Equal(t, "", o.FeastNames())
}

func TestOrthodoxCalSaintNames(t *testing.T) {
	o := new(OrthodoxCal)
	o.Saints = []string{"Prophet Micah", "Hieromartyr Pancratius"}
	assert.Equal(t, "Prophet Micah, Hieromartyr Pancratius", o.SaintNames())

	o.Saints = nil
	assert.Equal(t, "", o.SaintNames())
}
