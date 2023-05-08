package segments

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"

	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
)

type Owm struct {
	props properties.Properties
	env   platform.Environment

	Temperature int
	Weather     string
	URL         string
	units       string
	UnitIcon    string
}

const (
	// APIKey openweathermap api key
	APIKey properties.Property = "apikey"
	// Location openweathermap location
	Location properties.Property = "location"
	// Units openweathermap units
	Units properties.Property = "units"
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

func (d *Owm) Enabled() bool {
	err := d.setStatus()
	return err == nil
}

func (d *Owm) Template() string {
	return " {{ .Weather }} ({{ .Temperature }}{{ .UnitIcon }}) "
}

func (d *Owm) getResult() (*owmDataResponse, error) {
	cacheTimeout := d.props.GetInt(properties.CacheTimeout, properties.DefaultCacheTimeout)
	response := new(owmDataResponse)
	if cacheTimeout > 0 {
		// check if data stored in cache
		val, found := d.env.Cache().Get(CacheKeyResponse)
		// we got something from te cache
		if found {
			err := json.Unmarshal([]byte(val), response)
			if err != nil {
				return nil, err
			}
			d.URL, _ = d.env.Cache().Get(CacheKeyURL)
			return response, nil
		}
	}

	apikey := d.props.GetString(APIKey, ".")
	location := d.props.GetString(Location, "De Bilt,NL")
	units := d.props.GetString(Units, "standard")
	httpTimeout := d.props.GetInt(properties.HTTPTimeout, properties.DefaultHTTPTimeout)
	d.URL = fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?q=%s&units=%s&appid=%s", location, units, apikey)

	body, err := d.env.HTTPRequest(d.URL, nil, httpTimeout)
	if err != nil {
		return new(owmDataResponse), err
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return new(owmDataResponse), err
	}

	if cacheTimeout > 0 {
		// persist new forecasts in cache
		d.env.Cache().Set(CacheKeyResponse, string(body), cacheTimeout)
		d.env.Cache().Set(CacheKeyURL, d.URL, cacheTimeout)
	}
	return response, nil
}

func (d *Owm) setStatus() error {
	units := d.props.GetString(Units, "standard")
	q, err := d.getResult()
	if err != nil {
		return err
	}
	if len(q.Data) == 0 {
		return errors.New("No data found")
	}
	id := q.Data[0].TypeID

	d.Temperature = int(math.Round(q.temperature.Value))
	icon := ""
	switch id {
	case "01n":
		fallthrough
	case "01d":
		icon = "\U000f0599"
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
		icon = "\U000f0596"
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
	d.UnitIcon = "\ue33e"
	switch d.units {
	case "imperial":
		d.UnitIcon = "°F" // \ue341"
	case "metric":
		d.UnitIcon = "°C" // \ue339"
	case "":
		fallthrough
	case "standard":
		d.UnitIcon = "°K" // \U000f0506"
	}
	return nil
}

func (d *Owm) Init(props properties.Properties, env platform.Environment) {
	d.props = props
	d.env = env
}
