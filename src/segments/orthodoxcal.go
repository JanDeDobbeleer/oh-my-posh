package segments

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
)

// OrthodoxCal displays the Orthodox fasting level and feast information
// for the current day using the orthocal.info API.
type OrthodoxCal struct {
	Base

	orthodoxCalResponse
}

const (
	// OrthodoxCalType selects the calendar type: "gregorian" (default) or "julian".
	OrthodoxCalType options.Option = "calendar"
)

type orthodoxCalResponse struct {
	FastLevelDesc     string   `json:"fast_level_desc"`
	FastExceptionDesc string   `json:"fast_exception_desc"`
	SummaryTitle      string   `json:"summary_title"`
	FeastLevelDesc    string   `json:"feast_level_description"`
	Feasts            []string `json:"feasts"`
	Saints            []string `json:"saints"`
	Titles            []string `json:"titles"`
	FastLevel         int      `json:"fast_level"`
	FastException     int      `json:"fast_exception"`
	FeastLevel        int      `json:"feast_level"`
	Tone              int      `json:"tone"`
}

func (o *OrthodoxCal) Template() string {
	return " ☦ {{ .FastLevelDesc }}{{ if .FastExceptionDesc }} ({{ .FastExceptionDesc }}){{ end }} · {{ .SummaryTitle }} "
}

func (o *OrthodoxCal) Enabled() bool {
	if err := o.setData(); err != nil {
		log.Error(err)
		return false
	}

	return true
}

func (o *OrthodoxCal) setData() error {
	cal := o.options.String(OrthodoxCalType, "gregorian")
	url := fmt.Sprintf("https://orthocal.info/api/%s/", cal)

	httpTimeout := o.options.Int(options.HTTPTimeout, options.DefaultHTTPTimeout)

	body, err := o.env.HTTPRequest(url, nil, httpTimeout)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, &o.orthodoxCalResponse)
}

// IsFasting returns true when the day has any fasting requirement (fast_level > 0).
func (o *OrthodoxCal) IsFasting() bool {
	return o.FastLevel > 0
}

// FeastNames joins all feasts into a single comma-separated string.
func (o *OrthodoxCal) FeastNames() string {
	return strings.Join(o.Feasts, ", ")
}

// SaintNames joins all saints into a single comma-separated string.
func (o *OrthodoxCal) SaintNames() string {
	return strings.Join(o.Saints, ", ")
}
