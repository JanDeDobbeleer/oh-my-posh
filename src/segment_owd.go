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
}

const (
	APIKEY   Property = "apikey"
	LOCATION Property = "location"
	UNITS    Property = "units"
)

type weather struct {
	ShortDescription string `json:"main"`
	Description      string `json:"description"`
	TypeID           string `json:"icon"`
}
type temperature struct {
	Value float64 `json:"temp"`
}

type owmDataResponse struct {
	Data        []weather `json:"weather"`
	temperature `json:"main"`
}

func (d *owm) enabled() bool {
	err := d.setStatus()
	return err == nil
}

func (d *owm) string() string {
	return fmt.Sprintf("%s (%g°)", d.weather, d.temperature)
}

func (d *owm) setStatus() error {
	apikey := d.props.getString(APIKEY, ".")
	location := d.props.getString(LOCATION, "De Bilt,NL")
	units := d.props.getString(UNITS, "standard")

	url := fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?q=%s&units=%s&appid=%s", location, units, apikey)
	body, err := d.env.doGet(url)
	if err != nil {
		return err
	}
	q := new(owmDataResponse)
	err = json.Unmarshal(body, &q)
	if err != nil {
		return err
	}
	d.temperature = q.temperature.Value
	icon := ""
	switch q.Data[0].TypeID {
	case "01d":
		icon = "滛"
	case "02d":
		icon = "杖"
	case "03d":
		icon = "摒"
	case "04d":
		icon = ""
	case "09d":
		icon = "歹"
	case "10d":
		icon = "殺"
	case "11d":
		icon = "朗"
	case "13d":
		icon = ""
	case "50d":
		icon = ""
	}
	d.weather = icon
	return nil
}

func (d *owm) init(props *properties, env environmentInfo) {
	d.props = props
	d.env = env
}
