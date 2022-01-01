package main

import (
	"fmt"
	"testing"
	"time"

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

	cases := []struct {
		Case                      string
		BatchJSONResponse         string
		BatchReadingsJSONResponse string
		ExpectedString            string
		ExpectedEnabled           bool
		CacheTimeout              int
		CacheFoundFail            bool
		Template                  string
		Error                     error
	}{
		{
			Case: "Planning Status",
			BatchJSONResponse: `
			{"batchNo":18,"status":"Planning","brewDate":` + FakeBrewDateString + `,"bottlingDate":` + FakeBottlingDateString + `,"recipe":{"name":"Fake Beer"},"fermentationStartDate":` + FakeFermStartDateString + `,"name":"Batch","measuredAbv": 1.3}`, // nolint:lll
			BatchReadingsJSONResponse: `[]`,
			Template:                  "{{.StatusIcon}} {{if .DaysBottledOrFermented}}{{.DaysBottledOrFermented}}d {{end}}{{.Recipe.Name}} {{.MeasuredAbv}}%{{ if and (.Reading) (eq .Status \"Fermenting\")}}: {{.Reading.Gravity}} {{.Reading.Temperature}}° {{.TemperatureTrendIcon}}{{end}}", //nolint:lll
			ExpectedString:            " Fake Beer 1.3%",
			ExpectedEnabled:           true,
		},
		{
			Case: "Brewing Status",
			BatchJSONResponse: `
			{"batchNo":18,"status":"Brewing","brewDate":` + FakeBrewDateString + `,"bottlingDate":` + FakeBottlingDateString + `,"recipe":{"name":"Fake Beer"},"fermentationStartDate":` + FakeFermStartDateString + `,"name":"Batch","measuredAbv": 1.3}`, // nolint:lll
			BatchReadingsJSONResponse: `[]`,
			Template:                  "{{.StatusIcon}} {{if .DaysBottledOrFermented}}{{.DaysBottledOrFermented}}d {{end}}{{.Recipe.Name}} {{.MeasuredAbv}}%{{ if and (.Reading) (eq .Status \"Fermenting\")}}: {{.Reading.Gravity}} {{.Reading.Temperature}}° {{.TemperatureTrendIcon}}{{end}}", //nolint:lll
			ExpectedString:            " Fake Beer 1.3%",
			ExpectedEnabled:           true,
		},
		{
			Case: "Fermenting Status, no readings",
			BatchJSONResponse: `
			{"batchNo":18,"status":"Fermenting","brewDate":` + FakeBrewDateString + `,"bottlingDate":` + FakeBottlingDateString + `,"recipe":{"name":"Fake Beer"},"fermentationStartDate":` + FakeFermStartDateString + `,"name":"Batch","measuredAbv": 1.3}`, // nolint:lll
			BatchReadingsJSONResponse: `[]`,
			Template:                  "{{.StatusIcon}} {{if .DaysBottledOrFermented}}{{.DaysBottledOrFermented}}d {{end}}{{.Recipe.Name}} {{.MeasuredAbv}}%{{ if and (.Reading) (eq .Status \"Fermenting\")}}: {{.Reading.Gravity}} {{.Reading.Temperature}}° {{.TemperatureTrendIcon}}{{end}}", //nolint:lll
			ExpectedString:            " 19d Fake Beer 1.3%",
			ExpectedEnabled:           true,
		},
		{
			Case: "Fermenting Status, one reading",
			BatchJSONResponse: `
			{"batchNo":18,"status":"Fermenting","brewDate":` + FakeBrewDateString + `,"bottlingDate":` + FakeBottlingDateString + `,"recipe":{"name":"Fake Beer"},"fermentationStartDate":` + FakeFermStartDateString + `,"name":"Batch","measuredAbv": 1.3}`, // nolint:lll
			BatchReadingsJSONResponse: `[{"id":"manual","temp":19.5,"comment":"","sg":1.066,"time":` + FakeReading1DateString + `,"type":"manual"}]`,
			Template:                  "{{.StatusIcon}} {{if .DaysBottledOrFermented}}{{.DaysBottledOrFermented}}d {{end}}{{.Recipe.Name}} {{.MeasuredAbv}}%{{ if and (.Reading) (eq .Status \"Fermenting\")}}: {{.Reading.Gravity}} {{.Reading.Temperature}}° {{.TemperatureTrendIcon}}{{end}}", //nolint:lll
			ExpectedString:            " 19d Fake Beer 1.3%: 1.066 19.5° →",
			ExpectedEnabled:           true,
		},
		{
			Case: "Fermenting Status, two readings, temp trending up",
			BatchJSONResponse: `
			{"batchNo":18,"status":"Fermenting","brewDate":` + FakeBrewDateString + `,"bottlingDate":` + FakeBottlingDateString + `,"recipe":{"name":"Fake Beer"},"fermentationStartDate":` + FakeFermStartDateString + `,"name":"Batch","measuredAbv": 1.3}`, // nolint:lll
			BatchReadingsJSONResponse: `[{"id":"manual","temp":19.5,"comment":"","sg":1.066,"time":` + FakeReading1DateString + `,"type":"manual"}, {"id":"manual","temp":21,"comment":"","sg":1.063,"time":` + FakeReading2DateString + `,"type":"manual"}]`,                                    // nolint:lll
			Template:                  "{{.StatusIcon}} {{if .DaysBottledOrFermented}}{{.DaysBottledOrFermented}}d {{end}}{{.Recipe.Name}} {{.MeasuredAbv}}%{{ if and (.Reading) (eq .Status \"Fermenting\")}}: {{.Reading.Gravity}} {{.Reading.Temperature}}° {{.TemperatureTrendIcon}}{{end}}", //nolint:lll
			ExpectedString:            " 19d Fake Beer 1.3%: 1.063 21° ↗",
			ExpectedEnabled:           true,
		},
		{
			Case: "Fermenting Status, three readings, temp trending hard down, include age of most recent reading",
			BatchJSONResponse: `
			{"batchNo":18,"status":"Fermenting","brewDate":` + FakeBrewDateString + `,"bottlingDate":` + FakeBottlingDateString + `,"recipe":{"name":"Fake Beer"},"fermentationStartDate":` + FakeFermStartDateString + `,"name":"Batch","measuredAbv": 1.3}`, // nolint:lll
			BatchReadingsJSONResponse: `[{"id":"manual","temp":19.5,"comment":"","sg":1.066,"time":` + FakeReading1DateString + `,"type":"manual"}, {"id":"manual","temp":21,"comment":"","sg":1.063,"time":` + FakeReading2DateString + `,"type":"manual"}, {"id":"manual","temp":15,"comment":"","sg":1.050,"time":` + FakeReading3DateString + `,"type":"manual"}]`, // nolint:lll
			Template:                  "{{.StatusIcon}} {{.ReadingAge}} {{if .DaysBottledOrFermented}}{{.DaysBottledOrFermented}}d {{end}}{{.Recipe.Name}} {{.MeasuredAbv}}%{{ if and (.Reading) (eq .Status \"Fermenting\")}}: {{.Reading.Gravity}} {{.Reading.Temperature}}° {{.TemperatureTrendIcon}}{{end}}",                                                       //nolint:lll
			ExpectedString:            " 451 19d Fake Beer 1.3%: 1.05 15° ↓↓",
			ExpectedEnabled:           true,
		},
		{
			Case:                      "Bad batch json, readings fine",
			BatchJSONResponse:         ``,
			BatchReadingsJSONResponse: `[{"id":"manual","temp":19.5,"comment":"","sg":1.066,"time":` + FakeReading1DateString + `,"type":"manual"}, {"id":"manual","temp":21,"comment":"","sg":1.063,"time":` + FakeReading2DateString + `,"type":"manual"}, {"id":"manual","temp":15,"comment":"","sg":1.050,"time":` + FakeReading3DateString + `,"type":"manual"}]`, // nolint:lll
			Template:                  "{{.StatusIcon}} {{.ReadingAge}} {{if .DaysBottledOrFermented}}{{.DaysBottledOrFermented}}d {{end}}{{.Recipe.Name}} {{.MeasuredAbv}}%{{ if and (.Reading) (eq .Status \"Fermenting\")}}: {{.Reading.Gravity}} {{.Reading.Temperature}}° {{.TemperatureTrendIcon}}{{end}}",                                                       //nolint:lll
			ExpectedString:            "",
			ExpectedEnabled:           false,
		},
		{
			Case: "Conditioning Status",
			BatchJSONResponse: `
			{"batchNo":18,"status":"Conditioning","brewDate":` + FakeBrewDateString + `,"bottlingDate":` + FakeBottlingDateString + `,"recipe":{"name":"Fake Beer"},"fermentationStartDate":` + FakeFermStartDateString + `,"name":"Batch","measuredAbv": 1.3}`, // nolint:lll
			BatchReadingsJSONResponse: `[{"id":"manual","temp":19.5,"comment":"","sg":1.066,"time":` + FakeReading1DateString + `,"type":"manual"}, {"id":"manual","temp":21,"comment":"","sg":1.063,"time":` + FakeReading2DateString + `,"type":"manual"}, {"id":"manual","temp":15,"comment":"","sg":1.050,"time":` + FakeReading3DateString + `,"type":"manual"}]`, // nolint:lll
			Template:                  "{{.StatusIcon}} {{if .DaysBottledOrFermented}}{{.DaysBottledOrFermented}}d {{end}}{{.Recipe.Name}} {{.MeasuredAbv}}%{{ if and (.Reading) (eq .Status \"Fermenting\")}}: {{.Reading.Gravity}} {{.Reading.Temperature}}° {{.TemperatureTrendIcon}}{{end}}",                                                                       //nolint:lll
			ExpectedString:            " 5d Fake Beer 1.3%",
			ExpectedEnabled:           true,
		},
		{
			Case: "Fermenting Status, test all unit conversions",
			BatchJSONResponse: `
			{"batchNo":18,"status":"Fermenting","brewDate":` + FakeBrewDateString + `,"bottlingDate":` + FakeBottlingDateString + `,"recipe":{"name":"Fake Beer"},"fermentationStartDate":` + FakeFermStartDateString + `,"name":"Batch","measuredAbv": 1.3}`, // nolint:lll
			BatchReadingsJSONResponse: `[{"id":"manual","temp":34.5,"comment":"","sg":1.066,"time":` + FakeReading1DateString + `,"type":"manual"}]`,
			Template:                  "{{ if and (.Reading) (eq .Status \"Fermenting\") }}SG: ({{.Reading.Gravity}} Bx:{{.SGToBrix .Reading.Gravity}} P:{{.SGToPlato .Reading.Gravity}}), Temp: (C:{{.Reading.Temperature}} F:{{.DegCToF .Reading.Temperature}} K:{{.DegCToKelvin .Reading.Temperature}}){{end}}", //nolint:lll
			ExpectedString:            "SG: (1.066 Bx:16.13 P:16.13), Temp: (C:34.5 F:94.1 K:307.7)",
			ExpectedEnabled:           true,
		},
		{
			Case: "Fermenting Status, test all unit conversions 2",
			BatchJSONResponse: `
			{"batchNo":18,"status":"Fermenting","brewDate":` + FakeBrewDateString + `,"bottlingDate":` + FakeBottlingDateString + `,"recipe":{"name":"Fake Beer"},"fermentationStartDate":` + FakeFermStartDateString + `,"name":"Batch","measuredAbv": 1.3}`, // nolint:lll
			BatchReadingsJSONResponse: `[{"id":"manual","temp":3.5,"comment":"","sg":1.004,"time":` + FakeReading1DateString + `,"type":"manual"}]`,
			Template:                  "{{ if and (.Reading) (eq .Status \"Fermenting\") }}SG: ({{.Reading.Gravity}} Bx:{{.SGToBrix .Reading.Gravity}} P:{{.SGToPlato .Reading.Gravity}}), Temp: (C:{{.Reading.Temperature}} F:{{.DegCToF .Reading.Temperature}} K:{{.DegCToKelvin .Reading.Temperature}}){{end}}", //nolint:lll
			ExpectedString:            "SG: (1.004 Bx:1.03 P:1.03), Temp: (C:3.5 F:38.3 K:276.7)",
			ExpectedEnabled:           true,
		},
	}

	for _, tc := range cases {
		env := &MockedEnvironment{}
		props := properties{
			CacheTimeout: tc.CacheTimeout,
			BFBatchID:    BFFakeBatchID,
			BFAPIKey:     "FAKE",
			BFUserID:     "FAKE",
		}

		cache := &MockedCache{}
		cache.On("get", BFCacheKey).Return(nil, false) // cache testing later because cache is a little more complicated than just the single response.
		// cache.On("set", BFCacheKey, tc.JSONResponse, tc.CacheTimeout).Return()

		env.On("doGet", BFBatchURL).Return([]byte(tc.BatchJSONResponse), tc.Error)
		env.On("doGet", BFBatchReadingsURL).Return([]byte(tc.BatchReadingsJSONResponse), tc.Error)
		env.On("cache", nil).Return(cache)

		if tc.Template != "" {
			props[SegmentTemplate] = tc.Template
		}

		ns := &brewfather{
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
