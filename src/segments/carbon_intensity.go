package segments

import (
	"encoding/json"
	"fmt"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

type CarbonIntensity struct {
	props properties.Properties
	env   runtime.Environment

	TrendIcon string

	CarbonIntensityData
}

type CarbonIntensityResponse struct {
	Data []CarbonIntensityPeriod `json:"data"`
}

type CarbonIntensityPeriod struct {
	Intensity *CarbonIntensityData `json:"intensity"`
	From      string               `json:"from"`
	To        string               `json:"to"`
}

type CarbonIntensityData struct {
	Index    Index  `json:"index"`
	Forecast Number `json:"forecast"`
	Actual   Number `json:"actual"`
}

type Number int

func (n Number) String() string {
	if n == 0 {
		return "??"
	}

	return fmt.Sprintf("%d", n)
}

type Index string

func (i Index) Icon() string {
	switch i {
	case "very low":
		return "↓↓"
	case "low":
		return "↓"
	case "moderate":
		return "•"
	case "high":
		return "↑"
	case "very high":
		return "↑↑"
	default:
		return ""
	}
}

func (d *CarbonIntensity) Enabled() bool {
	err := d.setStatus()

	if err != nil {
		d.env.Error(err)
		return false
	}

	return true
}

func (d *CarbonIntensity) Template() string {
	return " CO₂ {{ .Index.Icon }}{{ .Actual.String }} {{ .TrendIcon }} {{ .Forecast.String }} "
}

func (d *CarbonIntensity) getResult() (*CarbonIntensityResponse, error) {
	response := new(CarbonIntensityResponse)
	url := "https://api.carbonintensity.org.uk/intensity"

	httpTimeout := d.props.GetInt(properties.HTTPTimeout, properties.DefaultHTTPTimeout)

	body, err := d.env.HTTPRequest(url, nil, httpTimeout)
	if err != nil {
		return new(CarbonIntensityResponse), err
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return new(CarbonIntensityResponse), err
	}

	return response, nil
}

func (d *CarbonIntensity) setStatus() error {
	response, err := d.getResult()
	if err != nil {
		return err
	}

	if len(response.Data) == 0 {
		d.Actual = 0
		d.Forecast = 0
		d.Index = "??"
		d.TrendIcon = "→"
		return nil
	}

	d.CarbonIntensityData = *response.Data[0].Intensity

	if d.Forecast > d.Actual {
		d.TrendIcon = "↗"
	}

	if d.Forecast < d.Actual {
		d.TrendIcon = "↘"
	}

	if d.Forecast == d.Actual || d.Actual == 0 || d.Forecast == 0 {
		d.TrendIcon = "→"
	}

	return nil
}

func (d *CarbonIntensity) Init(props properties.Properties, env runtime.Environment) {
	d.props = props
	d.env = env
}
