package main

import (
	"encoding/json"
	"fmt"
)

type owm struct {
	props       *properties
	env         environmentInfo
	temperature float64
	weather     string
	url         string
	units       string
}

const (
	// APIKEY openweathermap api key
	APIKEY Property = "apikey"
	// LOCATION openweathermap location
	LOCATION Property = "location"
	// UNITS openweathermap units
	UNITS Property = "units"
)

type weather struct {
	ShortDescription string `json:"main"`
	Description      string `json:"description"`
	TypeID           string `json:"icon"`
}
type temperature struct {
	Value float64 `json:"temp"`
}

type OWMDataResponse struct {
	Data        []weather `json:"weather"`
	temperature `json:"main"`
}

func (d *owm) enabled() bool {
	err := d.setStatus()
	return err == nil
}

func (d *owm) string() string {
	unitIcon := "\ue33e"
	switch d.units {
	case "imperial":
		unitIcon = "°F" // \ue341"
	case "metric":
		unitIcon = "°C" // \ue339"
	case "":
		fallthrough
	case "standard":
		unitIcon = "°K" // \ufa05"
	}
	text := fmt.Sprintf("%s (%g%s)", d.weather, d.temperature, unitIcon)
	if d.props.getBool(EnableHyperlink, false) {
		text = fmt.Sprintf("[%s](%s)", text, d.url)
	}
	return text
}

func (d *owm) setStatus() error {
	apikey := d.props.getString(APIKEY, ".")
	location := d.props.getString(LOCATION, "De Bilt,NL")
	units := d.props.getString(UNITS, "standard")
	timeout := d.props.getInt(HTTPTimeout, DefaultHTTPTimeout)

	url := fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?q=%s&units=%s&appid=%s", location, units, apikey)
	body, err := d.env.doGet(url, timeout)
	if err != nil {
		return err
	}
	q := new(OWMDataResponse)
	err = json.Unmarshal(body, &q)
	if err != nil {
		return err
	}
	d.temperature = q.temperature.Value
	icon := ""
	switch q.Data[0].TypeID {
	case "01d":
		icon = "\ufa98"
	case "02d":
		icon = "\ufa94"
	case "03d":
		icon = "\ue33d"
	case "04d":
		icon = "\ue312"
	case "09d":
		icon = "\ufa95"
	case "10d":
		icon = "\ue308"
	case "11d":
		icon = "\ue31d"
	case "13d":
		icon = "\ue31a"
	case "50d":
		icon = "\ue313"
	}
	d.weather = icon
	d.url = url
	d.units = units
	return nil
}

func (d *owm) init(props *properties, env environmentInfo) {
	d.props = props
	d.env = env
}
