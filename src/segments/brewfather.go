package segments

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"sort"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
)

// segment struct, makes templating easier
type Brewfather struct {
	Base

	DaysBottledOrFermented *uint
	TemperatureTrendIcon   string
	StatusIcon             string
	DayIcon                string
	URL                    string
	Batch
	ReadingAge     int
	DaysFermenting uint
	DaysBottled    uint
}

const (
	BFUserID  options.Option = "user_id"
	BFBatchID options.Option = "batch_id"

	BFDoubleUpIcon      options.Option = "doubleup_icon"
	BFSingleUpIcon      options.Option = "singleup_icon"
	BFFortyFiveUpIcon   options.Option = "fortyfiveup_icon"
	BFFlatIcon          options.Option = "flat_icon"
	BFFortyFiveDownIcon options.Option = "fortyfivedown_icon"
	BFSingleDownIcon    options.Option = "singledown_icon"
	BFDoubleDownIcon    options.Option = "doubledown_icon"

	BFPlanningStatusIcon     options.Option = "planning_status_icon"
	BFBrewingStatusIcon      options.Option = "brewing_status_icon"
	BFFermentingStatusIcon   options.Option = "fermenting_status_icon"
	BFConditioningStatusIcon options.Option = "conditioning_status_icon"
	BFCompletedStatusIcon    options.Option = "completed_status_icon"
	BFArchivedStatusIcon     options.Option = "archived_status_icon"

	BFDayIcon options.Option = "day_icon"

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
	DeviceType  string  `json:"type"`
	DeviceID    string  `json:"id"`
	Gravity     float64 `json:"sg"`
	Temperature float64 `json:"temp"`
	Timepoint   int64   `json:"timepoint"`
	Time        int64   `json:"time"`
}
type Batch struct {
	Reading   *BatchReading
	Status    string `json:"status"`
	BatchName string `json:"name"`
	Recipe    struct {
		Name string `json:"name"`
	} `json:"recipe"`
	BatchNumber      int     `json:"batchNo"`
	BrewDate         int64   `json:"brewDate"`
	FermentStartDate int64   `json:"fermentationStartDate"`
	BottlingDate     int64   `json:"bottlingDate"`
	MeasuredOg       float64 `json:"measuredOg"`
	MeasuredFg       float64 `json:"measuredFg"`
	MeasuredAbv      float64 `json:"measuredAbv"`
	TemperatureTrend float64
}

func (bf *Brewfather) Template() string {
	return " {{ .StatusIcon }} {{ if .DaysBottledOrFermented }}{{ .DaysBottledOrFermented }}{{ .DayIcon }} {{ end }}{{ url .Recipe.Name .URL }} {{ printf \"%.1f\" .MeasuredAbv }}%{{ if and (.Reading) (eq .Status \"Fermenting\") }} {{ printf \"%.3f\" .Reading.Gravity }} {{ .Reading.Temperature }}\u00b0 {{ .TemperatureTrendIcon }}{{ end }} " //nolint:lll
}

func (bf *Brewfather) Enabled() bool {
	data, err := bf.getResult()
	if err != nil {
		return false
	}
	bf.Batch = *data

	if bf.Reading != nil {
		readingDate := time.UnixMilli(bf.Reading.Time)
		bf.ReadingAge = int(time.Since(readingDate).Hours())
	} else {
		bf.ReadingAge = -1
	}

	bf.TemperatureTrendIcon = bf.getTrendIcon(bf.TemperatureTrend)
	bf.StatusIcon = bf.getBatchStatusIcon(data.Status)

	fermStartDate := time.UnixMilli(bf.FermentStartDate)
	bottlingDate := time.UnixMilli(bf.BottlingDate)

	switch bf.Status {
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
	batchID := bf.options.String(BFBatchID, "")
	if len(batchID) > 0 {
		bf.URL = fmt.Sprintf("https://web.brewfather.app/tabs/batches/batch/%s", batchID)
	}

	bf.DayIcon = bf.options.String(BFDayIcon, "d")

	return true
}

func (bf *Brewfather) getTrendIcon(trend float64) string {
	// Not a fan of this logic - wondering if Go lets us do something cleaner...
	if trend >= 0 {
		if trend > 4 {
			return bf.options.String(BFDoubleUpIcon, "↑↑")
		}

		if trend > 2 {
			return bf.options.String(BFSingleUpIcon, "↑")
		}

		if trend > 0.5 {
			return bf.options.String(BFFortyFiveUpIcon, "↗")
		}

		return bf.options.String(BFFlatIcon, "→")
	}

	if trend < -4 {
		return bf.options.String(BFDoubleDownIcon, "↓↓")
	}

	if trend < -2 {
		return bf.options.String(BFSingleDownIcon, "↓")
	}

	if trend < -0.5 {
		return bf.options.String(BFFortyFiveDownIcon, "↘")
	}

	return bf.options.String(BFFlatIcon, "→")
}

func (bf *Brewfather) getBatchStatusIcon(batchStatus string) string {
	switch batchStatus {
	case BFStatusPlanning:
		return bf.options.String(BFPlanningStatusIcon, "\uF8EA")
	case BFStatusBrewing:
		return bf.options.String(BFBrewingStatusIcon, "\uF7DE")
	case BFStatusFermenting:
		return bf.options.String(BFFermentingStatusIcon, "\uF499")
	case BFStatusConditioning:
		return bf.options.String(BFConditioningStatusIcon, "\uE372")
	case BFStatusCompleted:
		return bf.options.String(BFCompletedStatusIcon, "\uF7A5")
	case BFStatusArchived:
		return bf.options.String(BFArchivedStatusIcon, "\uF187")
	default:
		return ""
	}
}

func (bf *Brewfather) getResult() (*Batch, error) {
	userID := bf.options.String(BFUserID, "")
	if userID == "" {
		return nil, errors.New("missing Brewfather user id (user_id)")
	}

	apiKey := bf.options.String(APIKey, "")
	if apiKey == "" {
		return nil, errors.New("missing Brewfather api key (api_key)")
	}

	batchID := bf.options.String(BFBatchID, "")
	if batchID == "" {
		return nil, errors.New("missing Brewfather batch id (batch_id)")
	}

	authString := fmt.Sprintf("%s:%s", userID, apiKey)
	authStringb64 := base64.StdEncoding.EncodeToString([]byte(authString))
	authHeader := fmt.Sprintf("Basic %s", authStringb64)
	batchURL := fmt.Sprintf("https://api.brewfather.app/v1/batches/%s", batchID)
	batchReadingsURL := fmt.Sprintf("https://api.brewfather.app/v1/batches/%s/readings", batchID)

	httpTimeout := bf.options.Int(options.HTTPTimeout, options.DefaultHTTPTimeout)

	// batch
	addAuthHeader := func(request *http.Request) {
		request.Header.Add("authorization", authHeader)
	}

	body, err := bf.env.HTTPRequest(batchURL, nil, httpTimeout, addAuthHeader)
	if err != nil {
		return nil, err
	}

	var batch Batch
	err = json.Unmarshal(body, &batch)
	if err != nil {
		return nil, err
	}

	// readings
	body, err = bf.env.HTTPRequest(batchReadingsURL, nil, httpTimeout, addAuthHeader)
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
