package segments

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"oh-my-posh/environment"
	"oh-my-posh/properties"
	"sort"
	"time"
)

// segment struct, makes templating easier
type Brewfather struct {
	props properties.Properties
	env   environment.Environment

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
	BFUserID  properties.Property = "user_id"
	BFAPIKey  properties.Property = "api_key"
	BFBatchID properties.Property = "batch_id"

	BFDoubleUpIcon      properties.Property = "doubleup_icon"
	BFSingleUpIcon      properties.Property = "singleup_icon"
	BFFortyFiveUpIcon   properties.Property = "fortyfiveup_icon"
	BFFlatIcon          properties.Property = "flat_icon"
	BFFortyFiveDownIcon properties.Property = "fortyfivedown_icon"
	BFSingleDownIcon    properties.Property = "singledown_icon"
	BFDoubleDownIcon    properties.Property = "doubledown_icon"

	BFPlanningStatusIcon     properties.Property = "planning_status_icon"
	BFBrewingStatusIcon      properties.Property = "brewing_status_icon"
	BFFermentingStatusIcon   properties.Property = "fermenting_status_icon"
	BFConditioningStatusIcon properties.Property = "conditioning_status_icon"
	BFCompletedStatusIcon    properties.Property = "completed_status_icon"
	BFArchivedStatusIcon     properties.Property = "archived_status_icon"

	BFDayIcon properties.Property = "day_icon"

	BFCacheTimeout properties.Property = "cache_timeout"

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
	Status      string `json:"status"`
	BatchName   string `json:"name"`
	BatchNumber int    `json:"batchNo"`
	Recipe      struct {
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

func (bf *Brewfather) Template() string {
	return "{{ .StatusIcon }} {{ if .DaysBottledOrFermented }}{{ .DaysBottledOrFermented }}{{ .DayIcon }} {{ end }}{{ url .Recipe.Name .URL }} {{ printf \"%.1f\" .MeasuredAbv }}%{{ if and (.Reading) (eq .Status \"Fermenting\") }} {{ printf \"%.3f\" .Reading.Gravity }} {{ .Reading.Temperature }}\u00b0 {{ .TemperatureTrendIcon }}{{ end }}" // nolint:lll
}

func (bf *Brewfather) Enabled() bool {
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
	batchID := bf.props.GetString(BFBatchID, "")
	if len(batchID) > 0 {
		bf.URL = fmt.Sprintf("https://web.brewfather.app/tabs/batches/batch/%s", batchID)
	}

	bf.DayIcon = bf.props.GetString(BFDayIcon, "d")

	return true
}

func (bf *Brewfather) getTrendIcon(trend float64) string {
	// Not a fan of this logic - wondering if Go lets us do something cleaner...
	if trend >= 0 {
		if trend > 4 {
			return bf.props.GetString(BFDoubleUpIcon, "↑↑")
		}

		if trend > 2 {
			return bf.props.GetString(BFSingleUpIcon, "↑")
		}

		if trend > 0.5 {
			return bf.props.GetString(BFFortyFiveUpIcon, "↗")
		}

		return bf.props.GetString(BFFlatIcon, "→")
	}

	if trend < -4 {
		return bf.props.GetString(BFDoubleDownIcon, "↓↓")
	}

	if trend < -2 {
		return bf.props.GetString(BFSingleDownIcon, "↓")
	}

	if trend < -0.5 {
		return bf.props.GetString(BFFortyFiveDownIcon, "↘")
	}

	return bf.props.GetString(BFFlatIcon, "→")
}

func (bf *Brewfather) getBatchStatusIcon(batchStatus string) string {
	switch batchStatus {
	case BFStatusPlanning:
		return bf.props.GetString(BFPlanningStatusIcon, "\uF8EA")
	case BFStatusBrewing:
		return bf.props.GetString(BFBrewingStatusIcon, "\uF7DE")
	case BFStatusFermenting:
		return bf.props.GetString(BFFermentingStatusIcon, "\uF499")
	case BFStatusConditioning:
		return bf.props.GetString(BFConditioningStatusIcon, "\uE372")
	case BFStatusCompleted:
		return bf.props.GetString(BFCompletedStatusIcon, "\uF7A5")
	case BFStatusArchived:
		return bf.props.GetString(BFArchivedStatusIcon, "\uF187")
	default:
		return ""
	}
}

func (bf *Brewfather) getResult() (*Batch, error) {
	getFromCache := func(key string) (*Batch, error) {
		val, found := bf.env.Cache().Get(key)
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

		bf.env.Cache().Set(key, string(cacheJSON), cacheTimeout)

		return nil
	}

	userID := bf.props.GetString(BFUserID, "")
	if len(userID) == 0 {
		return nil, errors.New("missing Brewfather user id (user_id)")
	}

	apiKey := bf.props.GetString(BFAPIKey, "")
	if len(apiKey) == 0 {
		return nil, errors.New("missing Brewfather api key (api_key)")
	}

	batchID := bf.props.GetString(BFBatchID, "")
	if len(batchID) == 0 {
		return nil, errors.New("missing Brewfather batch id (batch_id)")
	}

	authString := fmt.Sprintf("%s:%s", userID, apiKey)
	authStringb64 := base64.StdEncoding.EncodeToString([]byte(authString))
	authHeader := fmt.Sprintf("Basic %s", authStringb64)
	batchURL := fmt.Sprintf("https://api.brewfather.app/v1/batches/%s", batchID)
	batchReadingsURL := fmt.Sprintf("https://api.brewfather.app/v1/batches/%s/readings", batchID)

	httpTimeout := bf.props.GetInt(HTTPTimeout, DefaultHTTPTimeout)
	cacheTimeout := bf.props.GetInt(BFCacheTimeout, 5)

	if cacheTimeout > 0 {
		if data, err := getFromCache(batchURL); err == nil {
			return data, nil
		}
	}

	// batch
	addAuthHeader := func(request *http.Request) {
		request.Header.Add("authorization", authHeader)
	}
	body, err := bf.env.HTTPRequest(batchURL, httpTimeout, addAuthHeader)
	if err != nil {
		return nil, err
	}

	var batch Batch
	err = json.Unmarshal(body, &batch)
	if err != nil {
		return nil, err
	}

	// readings
	body, err = bf.env.HTTPRequest(batchReadingsURL, httpTimeout, addAuthHeader)
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

// Unit conversion functions available to template.
func (bf *Brewfather) DegCToF(degreesC float64) float64 {
	return math.Round(10*((degreesC*1.8)+32)) / 10 // 1 decimal place
}

func (bf *Brewfather) DegCToKelvin(degreesC float64) float64 {
	return math.Round(10*(degreesC+273.15)) / 10 // 1 decimal place, only addition, but just to be sure
}

func (bf *Brewfather) SGToBrix(sg float64) float64 {
	// from https://en.wikipedia.org/wiki/Brix#Specific_gravity_2
	return math.Round(100*((182.4601*sg*sg*sg)-(775.6821*sg*sg)+(1262.7794*sg)-669.5622)) / 100
}

func (bf *Brewfather) SGToPlato(sg float64) float64 {
	// from https://en.wikipedia.org/wiki/Brix#Specific_gravity_2
	return math.Round(100*((135.997*sg*sg*sg)-(630.272*sg*sg)+(1111.14*sg)-616.868)) / 100 // 2 decimal places
}

func (bf *Brewfather) Init(props properties.Properties, env environment.Environment) {
	bf.props = props
	bf.env = env
}
