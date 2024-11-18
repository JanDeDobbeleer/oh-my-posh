package segments

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/stretchr/testify/assert"
)

const (
	BFFakeBatchID      = "FAKE"
	BFBatchURL         = "https://api.brewfather.app/v1/batches/" + BFFakeBatchID
	BFCacheKey         = BFBatchURL
	BFBatchReadingsURL = "https://api.brewfather.app/v1/batches/" + BFFakeBatchID + "/readings"
)

var (
	TimeNow = time.Now()
	// Create a fake timeline for the fake json, all in Unix milliseconds, to be used in all fake json responses
	FakeBrewDate              = TimeNow.Add(-time.Hour * 24 * 20)
	FakeFermentationStartDate = FakeBrewDate.Add(time.Hour * 24)                   // 1 day after brew date = 19 days ago
	FakeReading1Date          = FakeFermentationStartDate.Add(time.Minute * 35)    // first reading 35 minutes
	FakeReading2Date          = FakeReading1Date.Add(time.Hour)                    // second reading 1 hour later
	FakeReading3Date          = FakeReading2Date.Add(time.Hour * 3)                // 3 hours after last reading, 454 hours ago
	FakeBottlingDate          = FakeFermentationStartDate.Add(time.Hour * 24 * 14) // 14 days after ferm date = 5 days ago

	BrewDateMillis      = FakeBrewDate.UnixMilli()
	FermStartDateMillis = FakeFermentationStartDate.UnixMilli()
	Reading1DateMillis  = FakeReading1Date.UnixMilli()
	Reading2DateMillis  = FakeReading2Date.UnixMilli()
	Reading3DateMillis  = FakeReading3Date.UnixMilli()
	BottlingDateMillis  = FakeBottlingDate.UnixMilli()

	BatchNumber = 18
	BatchName   = "Batch"
	RecipeName  = "Fake Beer"
	MeasuredAbv = 1.3
)

func createBatch(status string) *Batch {
	return &Batch{
		Status:           status,
		BatchNumber:      BatchNumber,
		BrewDate:         BrewDateMillis,
		FermentStartDate: FermStartDateMillis,
		BottlingDate:     BottlingDateMillis,
		BatchName:        BatchName,
		MeasuredAbv:      MeasuredAbv,
		Recipe: struct {
			Name string `json:"name"`
		}{
			Name: RecipeName,
		},
	}
}

func createReading(temp, gravity float64, millis int64) *BatchReading {
	return &BatchReading{
		DeviceID:    "manual",
		Temperature: temp,
		Comment:     "",
		Gravity:     gravity,
		Time:        millis,
		DeviceType:  "manual",
	}
}

func TestBrewfatherSegment(t *testing.T) {
	cases := []struct {
		Error                 error
		BatchResponse         *Batch
		Case                  string
		ExpectedString        string
		Template              string
		BatchReadingsResponse []*BatchReading
		CacheTimeout          int
		ExpectedEnabled       bool
		CacheFoundFail        bool
	}{
		{
			Case:                  "Planning Status",
			BatchResponse:         createBatch(BFStatusPlanning),
			BatchReadingsResponse: []*BatchReading{},
			Template:              "{{.StatusIcon}} {{if .DaysBottledOrFermented}}{{.DaysBottledOrFermented}}d {{end}}{{.Recipe.Name}} {{.MeasuredAbv}}%{{ if and (.Reading) (eq .Status \"Fermenting\")}}: {{.Reading.Gravity}} {{.Reading.Temperature}}° {{.TemperatureTrendIcon}}{{end}}", //nolint:lll
			ExpectedString:        " Fake Beer 1.3%",
			ExpectedEnabled:       true,
		},
		{
			Case:                  "Brewing Status",
			BatchResponse:         createBatch(BFStatusBrewing),
			BatchReadingsResponse: []*BatchReading{},
			Template:              "{{.StatusIcon}} {{if .DaysBottledOrFermented}}{{.DaysBottledOrFermented}}d {{end}}{{.Recipe.Name}} {{.MeasuredAbv}}%{{ if and (.Reading) (eq .Status \"Fermenting\")}}: {{.Reading.Gravity}} {{.Reading.Temperature}}° {{.TemperatureTrendIcon}}{{end}}", //nolint:lll
			ExpectedString:        " Fake Beer 1.3%",
			ExpectedEnabled:       true,
		},
		{
			Case:                  "Fermenting Status, no readings",
			BatchResponse:         createBatch(BFStatusFermenting),
			BatchReadingsResponse: []*BatchReading{},
			Template:              "{{.StatusIcon}} {{if .DaysBottledOrFermented}}{{.DaysBottledOrFermented}}d {{end}}{{.Recipe.Name}} {{.MeasuredAbv}}%{{ if and (.Reading) (eq .Status \"Fermenting\")}}: {{.Reading.Gravity}} {{.Reading.Temperature}}° {{.TemperatureTrendIcon}}{{end}}", //nolint:lll
			ExpectedString:        " 19d Fake Beer 1.3%",
			ExpectedEnabled:       true,
		},
		{
			Case:          "Fermenting Status, one reading",
			BatchResponse: createBatch(BFStatusFermenting),
			BatchReadingsResponse: []*BatchReading{
				createReading(19.5, 1.066, Reading1DateMillis),
			},
			Template:        "{{.StatusIcon}} {{if .DaysBottledOrFermented}}{{.DaysBottledOrFermented}}d {{end}}{{.Recipe.Name}} {{.MeasuredAbv}}%{{ if and (.Reading) (eq .Status \"Fermenting\")}}: {{.Reading.Gravity}} {{.Reading.Temperature}}° {{.TemperatureTrendIcon}}{{end}}", //nolint:lll
			ExpectedString:  " 19d Fake Beer 1.3%: 1.066 19.5° →",
			ExpectedEnabled: true,
		},
		{
			Case:          "Fermenting Status, two readings, temp trending up",
			BatchResponse: createBatch(BFStatusFermenting),
			BatchReadingsResponse: []*BatchReading{
				createReading(21.0, 1.063, Reading2DateMillis),
				createReading(19.5, 1.066, Reading1DateMillis),
			},
			Template:        "{{.StatusIcon}} {{if .DaysBottledOrFermented}}{{.DaysBottledOrFermented}}d {{end}}{{.Recipe.Name}} {{.MeasuredAbv}}%{{ if and (.Reading) (eq .Status \"Fermenting\")}}: {{.Reading.Gravity}} {{.Reading.Temperature}}° {{.TemperatureTrendIcon}}{{end}}", //nolint:lll
			ExpectedString:  " 19d Fake Beer 1.3%: 1.063 21° ↗",
			ExpectedEnabled: true,
		},
		{
			Case:          "Fermenting Status, three readings, temp trending hard down, include age of most recent reading",
			BatchResponse: createBatch(BFStatusFermenting),
			BatchReadingsResponse: []*BatchReading{
				createReading(15.0, 1.050, Reading3DateMillis),
				createReading(21.0, 1.063, Reading2DateMillis),
				createReading(19.5, 1.066, Reading1DateMillis),
			},
			Template:        "{{.StatusIcon}} {{.ReadingAge}} {{if .DaysBottledOrFermented}}{{.DaysBottledOrFermented}}d {{end}}{{.Recipe.Name}} {{.MeasuredAbv}}%{{ if and (.Reading) (eq .Status \"Fermenting\")}}: {{.Reading.Gravity}} {{.Reading.Temperature}}° {{.TemperatureTrendIcon}}{{end}}", //nolint:lll
			ExpectedString:  " 451 19d Fake Beer 1.3%: 1.05 15° ↓↓",
			ExpectedEnabled: true,
		},
		{
			Case:          "Bad batch json, readings fine",
			BatchResponse: nil,
			BatchReadingsResponse: []*BatchReading{
				createReading(15.0, 1.050, Reading3DateMillis),
			},
			Template:        "{{.StatusIcon}} {{.ReadingAge}} {{if .DaysBottledOrFermented}}{{.DaysBottledOrFermented}}d {{end}}{{.Recipe.Name}} {{.MeasuredAbv}}%{{ if and (.Reading) (eq .Status \"Fermenting\")}}: {{.Reading.Gravity}} {{.Reading.Temperature}}° {{.TemperatureTrendIcon}}{{end}}", //nolint:lll
			ExpectedString:  "",
			ExpectedEnabled: false,
		},
		{
			Case:          "Conditioning Status",
			BatchResponse: createBatch(BFStatusConditioning),
			BatchReadingsResponse: []*BatchReading{
				createReading(15.0, 1.050, Reading3DateMillis),
			},
			Template:        "{{.StatusIcon}} {{if .DaysBottledOrFermented}}{{.DaysBottledOrFermented}}d {{end}}{{.Recipe.Name}} {{.MeasuredAbv}}%{{ if and (.Reading) (eq .Status \"Fermenting\")}}: {{.Reading.Gravity}} {{.Reading.Temperature}}° {{.TemperatureTrendIcon}}{{end}}", //nolint:lll
			ExpectedString:  " 5d Fake Beer 1.3%",
			ExpectedEnabled: true,
		},
		{
			Case:          "Fermenting Status, test all unit conversions",
			BatchResponse: createBatch(BFStatusFermenting),
			BatchReadingsResponse: []*BatchReading{
				createReading(34.5, 1.066, Reading1DateMillis),
			},
			Template:        "{{ if and (.Reading) (eq .Status \"Fermenting\") }}SG: ({{.Reading.Gravity}} Bx:{{.SGToBrix .Reading.Gravity}} P:{{.SGToPlato .Reading.Gravity}}), Temp: (C:{{.Reading.Temperature}} F:{{.DegCToF .Reading.Temperature}} K:{{.DegCToKelvin .Reading.Temperature}}){{end}}", //nolint:lll
			ExpectedString:  "SG: (1.066 Bx:16.13 P:16.13), Temp: (C:34.5 F:94.1 K:307.7)",
			ExpectedEnabled: true,
		},
		{
			Case:          "Fermenting Status, test all unit conversions 2",
			BatchResponse: createBatch(BFStatusFermenting),
			BatchReadingsResponse: []*BatchReading{
				createReading(3.5, 1.004, Reading1DateMillis),
			},
			Template:        "{{ if and (.Reading) (eq .Status \"Fermenting\") }}SG: ({{.Reading.Gravity}} Bx:{{.SGToBrix .Reading.Gravity}} P:{{.SGToPlato .Reading.Gravity}}), Temp: (C:{{.Reading.Temperature}} F:{{.DegCToF .Reading.Temperature}} K:{{.DegCToKelvin .Reading.Temperature}}){{end}}", //nolint:lll
			ExpectedString:  "SG: (1.004 Bx:1.03 P:1.03), Temp: (C:3.5 F:38.3 K:276.7)",
			ExpectedEnabled: true,
		},
	}

	for _, tc := range cases {
		env := &mock.Environment{}
		props := properties.Map{
			BFBatchID: BFFakeBatchID,
			APIKey:    "FAKE",
			BFUserID:  "FAKE",
		}

		var batchJSON []byte
		var err error
		if tc.BatchResponse != nil {
			batchJSON, err = json.Marshal(tc.BatchResponse)
			assert.NoError(t, err)
		} else {
			// bad JSON
			batchJSON = []byte("invalid json")
		}

		batchReadingsJSON, err := json.Marshal(tc.BatchReadingsResponse)
		assert.NoError(t, err)

		env.On("HTTPRequest", BFBatchURL).Return(batchJSON, tc.Error)
		env.On("HTTPRequest", BFBatchReadingsURL).Return(batchReadingsJSON, tc.Error)
		env.On("Flags").Return(&runtime.Flags{})

		brew := &Brewfather{}
		brew.Init(props, env)

		enabled := brew.Enabled()
		assert.Equal(t, tc.ExpectedEnabled, enabled, tc.Case)
		if !enabled {
			continue
		}

		if tc.Template == "" {
			tc.Template = brew.Template()
		}
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, tc.Template, brew), tc.Case)
	}
}
