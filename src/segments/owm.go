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
	APIKey properties.Property = "api_key"
	// Location openweathermap location
	Location properties.Property = "location"
	// Units openweathermap units
	Units properties.Property = "units"
	// Latitude for the location used in place of location
	Latitude properties.Property = "latitude"
	// Longitude for the location used in place of location
	Longitude properties.Property = "longitude"
	// CacheKeyResponse key used when caching the response
	CacheKeyResponse string = "owm_response"
	// CacheKeyURL key used when caching the url responsible for the response
	CacheKeyURL string = "owm_url"

	PoshOWMAPIKey = "POSH_OWM_API_KEY"
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

type geoLocation struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

func (d *Owm) Enabled() bool {
	err := d.setStatus()

	if err != nil {
		d.env.Error(err)
		return false
	}

	return true
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

	apikey := properties.OneOf[string](d.props, ".", APIKey, "apiKey")
	if len(apikey) == 0 {
		apikey = d.env.Getenv(PoshOWMAPIKey)
	}

	if len(apikey) == 0 {
		return nil, errors.New("no api key found")
	}

	location := d.props.GetString(Location, "De Bilt,NL")
	latitude := d.props.GetFloat64(Latitude, 91)    // This default value is intentionally invalid since there should not be a default for this and 0 is a valid value
	longitude := d.props.GetFloat64(Longitude, 181) // This default value is intentionally invalid since there should not be a default for this and 0 is a valid value
	units := d.props.GetString(Units, "standard")
	httpTimeout := d.props.GetInt(properties.HTTPTimeout, properties.DefaultHTTPTimeout)

	validCoordinates := func(latitude, longitude float64) bool {
		// Latitude values are only valid if they are between -90 and 90
		// Longitude values are only valid if they are between -180 and 180
		// https://gisgeography.com/latitude-longitude-coordinates/
		return latitude <= 90 && latitude >= -90 && longitude <= 180 && longitude >= -180
	}

	if !validCoordinates(latitude, longitude) {
		var geoResponse []geoLocation
		geocodingURL := fmt.Sprintf("http://api.openweathermap.org/geo/1.0/direct?q=%s&limit=1&appid=%s", location, apikey)

		body, err := d.env.HTTPRequest(geocodingURL, nil, httpTimeout)
		if err != nil {
			return new(owmDataResponse), err
		}

		err = json.Unmarshal(body, &geoResponse)
		if err != nil {
			return new(owmDataResponse), err
		}

		if len(geoResponse) == 0 {
			return new(owmDataResponse), fmt.Errorf("no coordinates found for %s", location)
		}

		latitude = geoResponse[0].Lat
		longitude = geoResponse[0].Lon
	}

	d.URL = fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?lat=%v&lon=%v&units=%s&appid=%s", latitude, longitude, units, apikey)

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
		icon = "\ue32b"
	case "01d":
		icon = "\ue30d"
	case "02n":
		icon = "\ue37e"
	case "02d":
		icon = "\ue302"
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
		icon = "\ue319"
	case "10n":
		icon = "\ue325"
	case "10d":
		icon = "\ue308"
	case "11n":
		icon = "\ue32a"
	case "11d":
		icon = "\ue30f"
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
		d.UnitIcon = "°K" // <b>K</b>"
	}
	return nil
}

func (d *Owm) Init(props properties.Properties, env platform.Environment) {
	d.props = props
	d.env = env
}
