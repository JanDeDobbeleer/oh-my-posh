package segments

import (
	"fmt"
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

func TestBrewfatherSegment(t *testing.T) {
	TimeNow := time.Now()

	// Create a fake timeline for the fake json, all in Unix milliseconds, to be used in all fake json responses
	FakeBrewDate := TimeNow.Add(-time.Hour * 24 * 20)
	FakeFermentationStartDate := FakeBrewDate.Add(time.Hour * 24)          // 1 day after brew date = 19 days ago
	FakeReading1Date := FakeFermentationStartDate.Add(time.Minute * 35)    // first reading 35 minutes
	FakeReading2Date := FakeReading1Date.Add(time.Hour)                    // second reading 1 hour later
	FakeReading3Date := FakeReading2Date.Add(time.Hour * 3)                // 3 hours after last reading, 454 hours ago
	FakeBottlingDate := FakeFermentationStartDate.Add(time.Hour * 24 * 14) // 14 days after ferm date = 5 days ago

	FakeBrewDateString := fmt.Sprintf("%d", FakeBrewDate.UnixMilli())
	FakeFermStartDateString := fmt.Sprintf("%d", FakeFermentationStartDate.UnixMilli())
	FakeReading1DateString := fmt.Sprintf("%d", FakeReading1Date.UnixMilli())
	FakeReading2DateString := fmt.Sprintf("%d", FakeReading2Date.UnixMilli())
	FakeReading3DateString := fmt.Sprintf("%d", FakeReading3Date.UnixMilli())
	FakeBottlingDateString := fmt.Sprintf("%d", FakeBottlingDate.UnixMilli())

	// TODO: make this smarter

	beerRecipeText := `,"recipe":{"name":"Fake Beer"},"fermentationStartDate":`
	bottlingDate := `,"bottlingDate":`
	batchText := `,"name":"Batch","measuredAbv": 1.3}`
	batchNumberFermenting := `{"batchNo":18,"status":"Fermenting","brewDate":`
	manualType := `,"type":"manual"}]`

	jsonFiller := FakeBrewDateString + bottlingDate + FakeBottlingDateString + beerRecipeText + FakeFermStartDateString + batchText

	batchReadingJSON := `[{"id":"manual","temp":19.5,"comment":"","sg":1.066,"time":`
	batchReadingJSON2 := `,"type":"manual"}, {"id":"manual","temp":21,"comment":"","sg":1.063,"time":`
	batchReadingJSON3 := `,"type":"manual"}, {"id":"manual","temp":15,"comment":"","sg":1.050,"time":`

	cases := []struct {
		Error                     error
		Case                      string
		BatchJSONResponse         string
		BatchReadingsJSONResponse string
		ExpectedString            string
		Template                  string
		CacheTimeout              int
		ExpectedEnabled           bool
		CacheFoundFail            bool
	}{
		{
			Case:                      "Planning Status",
			BatchJSONResponse:         `{"batchNo":18,"status":"Planning","brewDate":` + jsonFiller,
			BatchReadingsJSONResponse: `[]`,
			Template:                  "{{.StatusIcon}} {{if .DaysBottledOrFermented}}{{.DaysBottledOrFermented}}d {{end}}{{.Recipe.Name}} {{.MeasuredAbv}}%{{ if and (.Reading) (eq .Status \"Fermenting\")}}: {{.Reading.Gravity}} {{.Reading.Temperature}}° {{.TemperatureTrendIcon}}{{end}}", //nolint:lll
			ExpectedString:            " Fake Beer 1.3%",
			ExpectedEnabled:           true,
		},
		{
			Case:                      "Brewing Status",
			BatchJSONResponse:         `{"batchNo":18,"status":"Brewing","brewDate":` + jsonFiller,
			BatchReadingsJSONResponse: `[]`,
			Template:                  "{{.StatusIcon}} {{if .DaysBottledOrFermented}}{{.DaysBottledOrFermented}}d {{end}}{{.Recipe.Name}} {{.MeasuredAbv}}%{{ if and (.Reading) (eq .Status \"Fermenting\")}}: {{.Reading.Gravity}} {{.Reading.Temperature}}° {{.TemperatureTrendIcon}}{{end}}", //nolint:lll
			ExpectedString:            " Fake Beer 1.3%",
			ExpectedEnabled:           true,
		},
		{
			Case:                      "Fermenting Status, no readings",
			BatchJSONResponse:         batchNumberFermenting + jsonFiller,
			BatchReadingsJSONResponse: `[]`,
			Template:                  "{{.StatusIcon}} {{if .DaysBottledOrFermented}}{{.DaysBottledOrFermented}}d {{end}}{{.Recipe.Name}} {{.MeasuredAbv}}%{{ if and (.Reading) (eq .Status \"Fermenting\")}}: {{.Reading.Gravity}} {{.Reading.Temperature}}° {{.TemperatureTrendIcon}}{{end}}", //nolint:lll
			ExpectedString:            " 19d Fake Beer 1.3%",
			ExpectedEnabled:           true,
		},
		{
			Case:                      "Fermenting Status, one reading",
			BatchJSONResponse:         batchNumberFermenting + jsonFiller,
			BatchReadingsJSONResponse: batchReadingJSON + FakeReading1DateString + manualType,
			Template:                  "{{.StatusIcon}} {{if .DaysBottledOrFermented}}{{.DaysBottledOrFermented}}d {{end}}{{.Recipe.Name}} {{.MeasuredAbv}}%{{ if and (.Reading) (eq .Status \"Fermenting\")}}: {{.Reading.Gravity}} {{.Reading.Temperature}}° {{.TemperatureTrendIcon}}{{end}}", //nolint:lll
			ExpectedString:            " 19d Fake Beer 1.3%: 1.066 19.5° →",
			ExpectedEnabled:           true,
		},
		{
			Case:                      "Fermenting Status, two readings, temp trending up",
			BatchJSONResponse:         batchNumberFermenting + jsonFiller,                                                                                                                                                                                                                        //nolint:lll
			BatchReadingsJSONResponse: batchReadingJSON + FakeReading1DateString + batchReadingJSON2 + FakeReading2DateString + manualType,                                                                                                                                                       //nolint:lll
			Template:                  "{{.StatusIcon}} {{if .DaysBottledOrFermented}}{{.DaysBottledOrFermented}}d {{end}}{{.Recipe.Name}} {{.MeasuredAbv}}%{{ if and (.Reading) (eq .Status \"Fermenting\")}}: {{.Reading.Gravity}} {{.Reading.Temperature}}° {{.TemperatureTrendIcon}}{{end}}", //nolint:lll
			ExpectedString:            " 19d Fake Beer 1.3%: 1.063 21° ↗",
			ExpectedEnabled:           true,
		},
		{
			Case:                      "Fermenting Status, three readings, temp trending hard down, include age of most recent reading",
			BatchJSONResponse:         batchNumberFermenting + jsonFiller,                                                                                                                                                                                                                                        //nolint:lll
			BatchReadingsJSONResponse: batchReadingJSON + FakeReading1DateString + batchReadingJSON2 + FakeReading2DateString + batchReadingJSON3 + FakeReading3DateString + manualType,                                                                                                                          //nolint:lll
			Template:                  "{{.StatusIcon}} {{.ReadingAge}} {{if .DaysBottledOrFermented}}{{.DaysBottledOrFermented}}d {{end}}{{.Recipe.Name}} {{.MeasuredAbv}}%{{ if and (.Reading) (eq .Status \"Fermenting\")}}: {{.Reading.Gravity}} {{.Reading.Temperature}}° {{.TemperatureTrendIcon}}{{end}}", //nolint:lll
			ExpectedString:            " 451 19d Fake Beer 1.3%: 1.05 15° ↓↓",
			ExpectedEnabled:           true,
		},
		{
			Case:                      "Bad batch json, readings fine",
			BatchJSONResponse:         ``,
			BatchReadingsJSONResponse: batchReadingJSON + FakeReading1DateString + batchReadingJSON2 + FakeReading2DateString + batchReadingJSON3 + FakeReading3DateString + manualType,                                                                                                                          //nolint:lll
			Template:                  "{{.StatusIcon}} {{.ReadingAge}} {{if .DaysBottledOrFermented}}{{.DaysBottledOrFermented}}d {{end}}{{.Recipe.Name}} {{.MeasuredAbv}}%{{ if and (.Reading) (eq .Status \"Fermenting\")}}: {{.Reading.Gravity}} {{.Reading.Temperature}}° {{.TemperatureTrendIcon}}{{end}}", //nolint:lll
			ExpectedString:            "",
			ExpectedEnabled:           false,
		},
		{
			Case:                      "Conditioning Status",
			BatchJSONResponse:         `{"batchNo":18,"status":"Conditioning","brewDate":` + jsonFiller,                                                                                                                                                                                          //nolint:lll
			BatchReadingsJSONResponse: batchReadingJSON + FakeReading1DateString + batchReadingJSON2 + FakeReading2DateString + batchReadingJSON3 + FakeReading3DateString + manualType,                                                                                                          //nolint:lll
			Template:                  "{{.StatusIcon}} {{if .DaysBottledOrFermented}}{{.DaysBottledOrFermented}}d {{end}}{{.Recipe.Name}} {{.MeasuredAbv}}%{{ if and (.Reading) (eq .Status \"Fermenting\")}}: {{.Reading.Gravity}} {{.Reading.Temperature}}° {{.TemperatureTrendIcon}}{{end}}", //nolint:lll
			ExpectedString:            " 5d Fake Beer 1.3%",
			ExpectedEnabled:           true,
		},
		{
			Case:                      "Fermenting Status, test all unit conversions",
			BatchJSONResponse:         batchNumberFermenting + jsonFiller,
			BatchReadingsJSONResponse: `[{"id":"manual","temp":34.5,"comment":"","sg":1.066,"time":` + FakeReading1DateString + manualType,
			Template:                  "{{ if and (.Reading) (eq .Status \"Fermenting\") }}SG: ({{.Reading.Gravity}} Bx:{{.SGToBrix .Reading.Gravity}} P:{{.SGToPlato .Reading.Gravity}}), Temp: (C:{{.Reading.Temperature}} F:{{.DegCToF .Reading.Temperature}} K:{{.DegCToKelvin .Reading.Temperature}}){{end}}", //nolint:lll
			ExpectedString:            "SG: (1.066 Bx:16.13 P:16.13), Temp: (C:34.5 F:94.1 K:307.7)",
			ExpectedEnabled:           true,
		},
		{
			Case:                      "Fermenting Status, test all unit conversions 2",
			BatchJSONResponse:         batchNumberFermenting + jsonFiller,
			BatchReadingsJSONResponse: `[{"id":"manual","temp":3.5,"comment":"","sg":1.004,"time":` + FakeReading1DateString + manualType,
			Template:                  "{{ if and (.Reading) (eq .Status \"Fermenting\") }}SG: ({{.Reading.Gravity}} Bx:{{.SGToBrix .Reading.Gravity}} P:{{.SGToPlato .Reading.Gravity}}), Temp: (C:{{.Reading.Temperature}} F:{{.DegCToF .Reading.Temperature}} K:{{.DegCToKelvin .Reading.Temperature}}){{end}}", //nolint:lll
			ExpectedString:            "SG: (1.004 Bx:1.03 P:1.03), Temp: (C:3.5 F:38.3 K:276.7)",
			ExpectedEnabled:           true,
		},
	}

	for _, tc := range cases {
		env := &mock.Environment{}
		props := properties.Map{
			BFBatchID: BFFakeBatchID,
			APIKey:    "FAKE",
			BFUserID:  "FAKE",
		}

		env.On("HTTPRequest", BFBatchURL).Return([]byte(tc.BatchJSONResponse), tc.Error)
		env.On("HTTPRequest", BFBatchReadingsURL).Return([]byte(tc.BatchReadingsJSONResponse), tc.Error)
		env.On("Flags").Return(&runtime.Flags{})

		ns := &Brewfather{
			props: props,
			env:   env,
		}

		enabled := ns.Enabled()
		assert.Equal(t, tc.ExpectedEnabled, enabled, tc.Case)
		if !enabled {
			continue
		}

		if tc.Template == "" {
			tc.Template = ns.Template()
		}
		assert.Equal(t, tc.ExpectedString, renderTemplate(env, tc.Template, ns), tc.Case)
	}
}
