package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"time"
)

// segment struct, makes templating easier
type brewfather struct {
	props properties
	env   environmentInfo

	Batch
	TemperatureTrendIcon string
	StatusIcon           string
	DayIcon              string // populated from day_icon for use in template

	ReadingAge             int // age in hours of the most recent reading included in the batch, -1 if none
	DaysFermenting         uint
	DaysBottled            uint
	DaysBottledOrFermented *uint // help avoid chronic template logic - code will point this to one of above or be nil depending on status

	URL string // URL of batch page to open if hyperlink enabled on the segment and URL formatting used in template: [name](link)
}

const (
	BFUserID  Property = "user_id"
	BFAPIKey  Property = "api_key"
	BFBatchID Property = "batch_id"

	BFDoubleUpIcon      Property = "doubleup_icon"
	BFSingleUpIcon      Property = "singleup_icon"
	BFFortyFiveUpIcon   Property = "fortyfiveup_icon"
	BFFlatIcon          Property = "flat_icon"
	BFFortyFiveDownIcon Property = "fortyfivedown_icon"
	BFSingleDownIcon    Property = "singledown_icon"
	BFDoubleDownIcon    Property = "doubledown_icon"

	BFPlanningStatusIcon     Property = "planning_status_icon"
	BFBrewingStatusIcon      Property = "brewing_status_icon"
	BFFermentingStatusIcon   Property = "fermenting_status_icon"
	BFConditioningStatusIcon Property = "conditioning_status_icon"
	BFCompletedStatusIcon    Property = "completed_status_icon"
	BFArchivedStatusIcon     Property = "archived_status_icon"

	BFDayIcon Property = "day_icon"

	BFCacheTimeout Property = "cache_timeout"

	DefaultTemplate string = "{{.StatusIcon}} {{if .DaysBottledOrFermented}}{{.DaysBottledOrFermented}}{{.DayIcon}} {{end}}[{{.Recipe.Name}}]({{.URL}})" +
		" {{printf \"%.1f\" .MeasuredAbv}}%{{ if and (.Reading) (eq .Status \"Fermenting\")}} " +
		"{{printf \"%.3f\" .Reading.Gravity}} {{.Reading.Temperature}}\u00b0 {{.TemperatureTrendIcon}}{{end}}"

	BFStatusPlanning     string = "Planning"
	BFStatusBrewing      string = "Brewing"
	BFStatusFermenting   string = "Fermenting"
	BFStatusConditioning string = "Conditioning"
	BFStatusCompleted    string = "Completed"
	BFStatusArchived     string = "Archived"
)

// Returned from https://api.brewfather.app/v1/batches/batch_id/readings
type BatchReading struct {
	Comment     string  `json:"comment"`
	Gravity     float64 `json:"sg"`
	DeviceType  string  `json:"type"`
	DeviceID    string  `json:"id"`
	Temperature float64 `json:"temp"`      // celsius - need to add F conversion
	Timepoint   int64   `json:"timepoint"` // << check what these are...
	Time        int64   `json:"time"`      // <<
}
type Batch struct {
	// Json tagged values returned from https://api.brewfather.app/v1/batches/batch_id
	Status string `json:"status"`
	Recipe struct {
		Name string `json:"name"`
	} `json:"recipe"`
	BrewDate         int64 `json:"brewDate"`
	FermentStartDate int64 `json:"fermentationStartDate"`
	BottlingDate     int64 `json:"bottlingDate"`

	MeasuredOg  float64 `json:"measuredOg"`
	MeasuredFg  float64 `json:"measuredFg"`
	MeasuredAbv float64 `json:"measuredAbv"`

	// copy of the latest BatchReading in here.
	Reading *BatchReading

	// Calculated values we need to cache because they require the rest query to reproduce
	TemperatureTrend float64 // diff between this and last, short term trend
}

func (bf *brewfather) enabled() bool {
	data, err := bf.getResult()
	if err != nil {
		return false
	}
	bf.Batch = *data

	if bf.Batch.Reading != nil {
		readingDate := time.UnixMilli(bf.Batch.Reading.Time)
		bf.ReadingAge = int(time.Since(readingDate).Hours())
	} else {
		bf.ReadingAge = -1
	}

	bf.TemperatureTrendIcon = bf.getTrendIcon(bf.TemperatureTrend)
	bf.StatusIcon = bf.getBatchStatusIcon(data.Status)

	fermStartDate := time.UnixMilli(bf.Batch.FermentStartDate)
	bottlingDate := time.UnixMilli(bf.Batch.BottlingDate)

	switch bf.Batch.Status {
	case BFStatusFermenting:
		// in the fermenter now, so relative to today.
		bf.DaysFermenting = uint(time.Since(fermStartDate).Hours() / 24)
		bf.DaysBottled = 0
		bf.DaysBottledOrFermented = &bf.DaysFermenting
	case BFStatusConditioning, BFStatusCompleted, BFStatusArchived:
		bf.DaysFermenting = uint(bottlingDate.Sub(fermStartDate).Hours() / 24)
		bf.DaysBottled = uint(time.Since(bottlingDate).Hours() / 24)
		bf.DaysBottledOrFermented = &bf.DaysBottled
	default:
		bf.DaysFermenting = 0
		bf.DaysBottled = 0
		bf.DaysBottledOrFermented = nil
	}

	// URL property set to weblink to the full batch page
	batchID := bf.props.getString(BFBatchID, "")
	if len(batchID) > 0 {
		bf.URL = fmt.Sprintf("https://web.brewfather.app/tabs/batches/batch/%s", batchID)
	}

	bf.DayIcon = bf.props.getString(BFDayIcon, "d")

	return true
}

func (bf *brewfather) getTrendIcon(trend float64) string {
	// Not a fan of this logic - wondering if Go lets us do something cleaner...
	if trend >= 0 {
		if trend > 4 {
			return bf.props.getString(BFDoubleUpIcon, "↑↑")
		}

		if trend > 2 {
			return bf.props.getString(BFSingleUpIcon, "↑")
		}

		if trend > 0.5 {
			return bf.props.getString(BFFortyFiveUpIcon, "↗")
		}

		return bf.props.getString(BFFlatIcon, "→")
	}

	if trend < -4 {
		return bf.props.getString(BFDoubleDownIcon, "↓↓")
	}

	if trend < -2 {
		return bf.props.getString(BFSingleDownIcon, "↓")
	}

	if trend < -0.5 {
		return bf.props.getString(BFFortyFiveDownIcon, "↘")
	}

	return bf.props.getString(BFFlatIcon, "→")
}

func (bf *brewfather) getBatchStatusIcon(batchStatus string) string {
	switch batchStatus {
	case BFStatusPlanning:
		return bf.props.getString(BFPlanningStatusIcon, "\uF8EA")
	case BFStatusBrewing:
		return bf.props.getString(BFBrewingStatusIcon, "\uF7DE")
	case BFStatusFermenting:
		return bf.props.getString(BFFermentingStatusIcon, "\uF499")
	case BFStatusConditioning:
		return bf.props.getString(BFConditioningStatusIcon, "\uE372")
	case BFStatusCompleted:
		return bf.props.getString(BFCompletedStatusIcon, "\uF7A5")
	case BFStatusArchived:
		return bf.props.getString(BFArchivedStatusIcon, "\uF187")
	default:
		return ""
	}
}

func (bf *brewfather) string() string {
	segmentTemplate := bf.props.getString(SegmentTemplate, DefaultTemplate)
	template := &textTemplate{
		Template: segmentTemplate,
		Context:  bf,
		Env:      bf.env,
	}
	text, err := template.render()
	if err != nil {
		return err.Error()
	}

	return text
}

func (bf *brewfather) getResult() (*Batch, error) {
	getFromCache := func(key string) (*Batch, error) {
		val, found := bf.env.cache().get(key)
		// we got something from the cache
		if found {
			var result Batch
			err := json.Unmarshal([]byte(val), &result)
			if err == nil {
				return &result, nil
			}
		}
		return nil, errors.New("no data in cache")
	}

	putToCache := func(key string, batch *Batch, cacheTimeout int) error {
		cacheJSON, err := json.Marshal(batch)
		if err != nil {
			return err
		}

		bf.env.cache().set(key, string(cacheJSON), cacheTimeout)

		return nil
	}

	userID := bf.props.getString(BFUserID, "")
	if len(userID) == 0 {
		return nil, errors.New("missing Brewfather user id (user_id)")
	}

	apiKey := bf.props.getString(BFAPIKey, "")
	if len(apiKey) == 0 {
		return nil, errors.New("missing Brewfather api key (api_key)")
	}

	batchID := bf.props.getString(BFBatchID, "")
	if len(batchID) == 0 {
		return nil, errors.New("missing Brewfather batch id (batch_id)")
	}

	authString := fmt.Sprintf("%s:%s", userID, apiKey)
	authStringb64 := base64.StdEncoding.EncodeToString([]byte(authString))
	authHeader := fmt.Sprintf("Basic %s", authStringb64)
	batchURL := fmt.Sprintf("https://api.brewfather.app/v1/batches/%s", batchID)
	batchReadingsURL := fmt.Sprintf("https://api.brewfather.app/v1/batches/%s/readings", batchID)

	httpTimeout := bf.props.getInt(HTTPTimeout, DefaultHTTPTimeout)
	cacheTimeout := bf.props.getInt(BFCacheTimeout, 5)

	if cacheTimeout > 0 {
		if data, err := getFromCache(batchURL); err == nil {
			return data, nil
		}
	}

	// batch
	addAuthHeader := func(request *http.Request) {
		request.Header.Add("authorization", authHeader)
	}
	body, err := bf.env.doGet(batchURL, httpTimeout, addAuthHeader)
	if err != nil {
		return nil, err
	}

	var batch Batch
	err = json.Unmarshal(body, &batch)
	if err != nil {
		return nil, err
	}

	// readings
	body, err = bf.env.doGet(batchReadingsURL, httpTimeout, addAuthHeader)
	if err != nil {
		return nil, err
	}

	var arr []*BatchReading
	err = json.Unmarshal(body, &arr)
	if err != nil {
		return nil, err
	}

	if len(arr) > 0 {
		// could just take latest reading using their API, but that won't allow us to see trend - get 'em all and sort by time,
		// using two most recent for trend
		sort.Slice(arr, func(i, j int) bool {
			return arr[i].Time > arr[j].Time
		})

		// Keep the latest one
		batch.Reading = arr[0]

		if len(arr) > 1 {
			batch.TemperatureTrend = arr[0].Temperature - arr[1].Temperature
		}
	}

	if cacheTimeout > 0 {
		_ = putToCache(batchURL, &batch, cacheTimeout)
	}

	return &batch, nil
}

func (bf *brewfather) init(props properties, env environmentInfo) {
	bf.props = props
	bf.env = env
}
