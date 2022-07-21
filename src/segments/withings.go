package segments

import (
	"errors"
	"fmt"
	"math"
	"oh-my-posh/environment"
	"oh-my-posh/http"
	"oh-my-posh/properties"
	"strconv"
	"strings"
	"time"

	http2 "net/http"
	"net/url"
)

// WithingsData struct contains the API data
type WithingsData struct {
	Status int   `json:"status"`
	Body   *Body `json:"body"`
}

type Body struct {
	MeasureGroups []*MeasureGroup `json:"measuregrps"`
	Activities    []*Activity     `json:"activities"`
	Series        []*Series       `json:"series"`
}

type MeasureGroup struct {
	Measures []*Measure  `json:"measures"`
	Comment  interface{} `json:"comment"`
}

type Measure struct {
	Value int `json:"value"`
	Type  int `json:"type"`
	Unit  int `json:"unit"`
}

type Series struct {
	Startdate int64 `json:"startdate"`
	Enddate   int64 `json:"enddate"`
}

type Activity struct {
	Date          string `json:"date"`
	Timezone      string `json:"timezone"`
	Deviceid      string `json:"deviceid"`
	HashDeviceid  string `json:"hash_deviceid"`
	Brand         int    `json:"brand"`
	IsTracker     bool   `json:"is_tracker"`
	Steps         int    `json:"steps"`
	Distance      int    `json:"distance"`
	Elevation     int    `json:"elevation"`
	Soft          int    `json:"soft"`
	Moderate      int    `json:"moderate"`
	Intense       int    `json:"intense"`
	Active        int    `json:"active"`
	Calories      int    `json:"calories"`
	Totalcalories int    `json:"totalcalories"`
	HrAverage     int    `json:"hr_average"`
	HrMin         int    `json:"hr_min"`
	HrMax         int    `json:"hr_max"`
	HrZone0       int    `json:"hr_zone_0"`
	HrZone1       int    `json:"hr_zone_1"`
	HrZone2       int    `json:"hr_zone_2"`
	HrZone3       int    `json:"hr_zone_3"`
}

// WithingsAPI is a wrapper around http.Oauth
type WithingsAPI interface {
	GetMeasures(meastypes string) (*WithingsData, error)
	GetActivities(activities string) (*WithingsData, error)
	GetSleep() (*WithingsData, error)
}

type withingsAPI struct {
	*http.OAuth
}

func (w *withingsAPI) GetMeasures(meastypes string) (*WithingsData, error) {
	twoWeeksAgo := strconv.FormatInt(time.Now().AddDate(0, 0, -14).Unix(), 10)
	formData := url.Values{
		"meastypes":  {meastypes},
		"action":     {"getmeas"},
		"lastupdate": {twoWeeksAgo},
		"category":   {"1"},
	}
	return w.getWithingsData("https://wbsapi.withings.net/measure", formData)
}

func (w *withingsAPI) GetActivities(activities string) (*WithingsData, error) {
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	today := time.Now().Format("2006-01-02")
	formData := url.Values{
		"data_fields":  {activities},
		"action":       {"getactivity"},
		"startdateymd": {yesterday},
		"enddateymd":   {today},
		"category":     {"1"},
	}
	return w.getWithingsData("https://wbsapi.withings.net/v2/measure", formData)
}

func (w *withingsAPI) GetSleep() (*WithingsData, error) {
	now := time.Now()
	formData := url.Values{
		"action":    {"get"},
		"startdate": {strconv.FormatInt(now.AddDate(0, -1, 0).Unix(), 10)},
		"enddate":   {strconv.FormatInt(now.Unix(), 10)},
	}
	return w.getWithingsData("https://wbsapi.withings.net/v2/sleep", formData)
}

func (w *withingsAPI) getWithingsData(endpoint string, formData url.Values) (*WithingsData, error) {
	modifiers := func(request *http2.Request) {
		request.Method = http2.MethodPost
		request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}
	body := strings.NewReader(formData.Encode())
	data, err := http.OauthResult[*WithingsData](w.OAuth, endpoint, body, modifiers)
	if data != nil && data.Status != 0 {
		return nil, errors.New("Withings API error: " + strconv.Itoa(data.Status))
	}
	return data, err
}

type Withings struct {
	props properties.Properties

	Weight     float64
	SleepHours string
	Steps      int

	api WithingsAPI
}

const (
	WithingsAccessTokenKey  = "withings_access_token"
	WithingsRefreshTokenKey = "withings_refresh_token"
)

func (w *Withings) Template() string {
	return "{{ if gt .Weight 0.0 }} {{ round .Weight 2 }}kg {{ end }}"
}

func (w *Withings) Enabled() bool {
	var enabled bool
	if w.getActivities() {
		enabled = true
	}
	if w.getMeasures() {
		enabled = true
	}
	if w.getSleep() {
		enabled = true
	}
	return enabled
}

func (w *Withings) getMeasures() bool {
	data, err := w.api.GetMeasures("1")
	if err != nil {
		return false
	}
	// no data
	if len(data.Body.MeasureGroups) == 0 || len(data.Body.MeasureGroups[0].Measures) == 0 {
		return false
	}
	measure := data.Body.MeasureGroups[0].Measures[0]
	weight := measure.Value
	w.Weight = float64(weight) / math.Pow(10, math.Abs(float64(measure.Unit)))
	return true
}

func (w *Withings) getActivities() bool {
	data, err := w.api.GetActivities("steps")
	if err != nil || len(data.Body.Activities) == 0 {
		return false
	}
	w.Steps = data.Body.Activities[0].Steps
	return true
}

func (w *Withings) getSleep() bool {
	data, err := w.api.GetSleep()
	if err != nil || len(data.Body.Series) == 0 {
		return false
	}
	sleepStart := time.Unix(data.Body.Series[0].Startdate, 0)
	sleepEnd := time.Unix(data.Body.Series[0].Enddate, 0)
	sleepHours := sleepEnd.Sub(sleepStart).Hours()
	w.SleepHours = fmt.Sprintf("%0.1f", sleepHours)
	return true
}

func (w *Withings) Init(props properties.Properties, env environment.Environment) {
	w.props = props

	w.api = &withingsAPI{
		OAuth: &http.OAuth{
			Props:           props,
			Env:             env,
			AccessTokenKey:  WithingsAccessTokenKey,
			RefreshTokenKey: WithingsRefreshTokenKey,
			SegmentName:     "withings",
		},
	}
}
