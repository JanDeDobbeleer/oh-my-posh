package main

import (
	"encoding/json"
	"fmt"
)

type owm struct {
	props       *properties
	env         environmentInfo
	Temperature float64
	Weather     string
	URL         string
	units       string
	UnitIcon    string
}

const (
	// APIKey openweathermap api key
	APIKey Property = "apikey"
	// Location openweathermap location
	Location Property = "location"
	// Units openweathermap units
	Units Property = "units"
	// CacheTimeout cache timeout
	CacheTimeout Property = "cache_timeout"
	// CacheKeyResponse key used when caching the response
	CacheKeyResponse string = "owm_response"
	// CacheKeyURL key used when caching the url responsible for the response
	CacheKeyURL string = "owm_url"
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
	d.UnitIcon = "\ue33e"
	switch d.units {
	case "imperial":
		d.UnitIcon = "°F" // \ue341"
	case "metric":
		d.UnitIcon = "°C" // \ue339"
	case "":
		fallthrough
	case "standard":
		d.UnitIcon = "°K" // \ufa05"
	}
	segmentTemplate := d.props.getString(SegmentTemplate, "{{.Weather}} ({{.Temperature}}{{.UnitIcon}})")
	template := &textTemplate{
		Template: segmentTemplate,
		Context:  d,
		Env:      d.env,
	}
	text, err := template.render()
	if err != nil {
		return err.Error()
	}

	return text
}

func (d *owm) getResult() (*owmDataResponse, error) {
	cacheTimeout := d.props.getInt(CacheTimeout, DefaultCacheTimeout)
	response := new(owmDataResponse)
	if cacheTimeout > 0 {
		// check if data stored in cache
		val, found := d.env.cache().get(CacheKeyResponse)
		// we got something from te cache
		if found {
			err := json.Unmarshal([]byte(val), response)
			if err != nil {
				return nil, err
			}
			d.URL, _ = d.env.cache().get(CacheKeyURL)
			return response, nil
		}
	}

	apikey := d.props.getString(APIKey, ".")
	location := d.props.getString(Location, "De Bilt,NL")
	units := d.props.getString(Units, "standard")
	httpTimeout := d.props.getInt(HTTPTimeout, DefaultHTTPTimeout)
	d.URL = fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?q=%s&units=%s&appid=%s", location, units, apikey)

	body, err := d.env.doGet(d.URL, httpTimeout)
	if err != nil {
		return new(owmDataResponse), err
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return new(owmDataResponse), err
	}

	if cacheTimeout > 0 {
		// persist new forecasts in cache
		d.env.cache().set(CacheKeyResponse, string(body), cacheTimeout)
		d.env.cache().set(CacheKeyURL, d.URL, cacheTimeout)
	}
	return response, nil
}

func (d *owm) setStatus() error {
	units := d.props.getString(Units, "standard")

	q, err := d.getResult()
	if err != nil {
		return err
	}

	d.Temperature = q.temperature.Value
	icon := ""
	switch q.Data[0].TypeID {
	case "01n":
		fallthrough
	case "01d":
		icon = "\ufa98"
	case "02n":
		fallthrough
	case "02d":
		icon = "\ufa94"
	case "03n":
		fallthrough
	case "03d":
		icon = "\ue33d"
	case "04n":
		fallthrough
	case "04d":
		icon = "\ue312"
	case "09n":
		fallthrough
	case "09d":
		icon = "\ufa95"
	case "10n":
		fallthrough
	case "10d":
		icon = "\ue308"
	case "11n":
		fallthrough
	case "11d":
		icon = "\ue31d"
	case "13n":
		fallthrough
	case "13d":
		icon = "\ue31a"
	case "50n":
		fallthrough
	case "50d":
		icon = "\ue313"
	}
	d.Weather = icon
	d.units = units
	return nil
}

func (d *owm) init(props *properties, env environmentInfo) {
	d.props = props
	d.env = env
}
