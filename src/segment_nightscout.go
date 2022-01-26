package main

import (
	"encoding/json"
	"errors"
	"oh-my-posh/environment"
	"oh-my-posh/properties"
	"time"
)

// segment struct, makes templating easier
type nightscout struct {
	props properties.Properties
	env   environment.Environment

	NightscoutData
	TrendIcon string
}

const (
	// Your complete Nightscout URL and APIKey like this
	URL properties.Property = "url"

	DoubleUpIcon      properties.Property = "doubleup_icon"
	SingleUpIcon      properties.Property = "singleup_icon"
	FortyFiveUpIcon   properties.Property = "fortyfiveup_icon"
	FlatIcon          properties.Property = "flat_icon"
	FortyFiveDownIcon properties.Property = "fortyfivedown_icon"
	SingleDownIcon    properties.Property = "singledown_icon"
	DoubleDownIcon    properties.Property = "doubledown_icon"

	NSCacheTimeout properties.Property = "cache_timeout"
)

// NightscoutData struct contains the API data
type NightscoutData struct {
	ID         string    `json:"_id"`
	Sgv        int       `json:"sgv"`
	Date       int64     `json:"date"`
	DateString time.Time `json:"dateString"`
	Trend      int       `json:"trend"`
	Direction  string    `json:"direction"`
	Device     string    `json:"device"`
	Type       string    `json:"type"`
	UtcOffset  int       `json:"utcOffset"`
	SysTime    time.Time `json:"sysTime"`
	Mills      int64     `json:"mills"`
}

func (ns *nightscout) template() string {
	return "{{ .Sgv }}"
}

func (ns *nightscout) enabled() bool {
	data, err := ns.getResult()
	if err != nil {
		return false
	}
	ns.NightscoutData = *data
	ns.TrendIcon = ns.getTrendIcon()

	return true
}

func (ns *nightscout) getTrendIcon() string {
	switch ns.Direction {
	case "DoubleUp":
		return ns.props.GetString(DoubleUpIcon, "↑↑")
	case "SingleUp":
		return ns.props.GetString(SingleUpIcon, "↑")
	case "FortyFiveUp":
		return ns.props.GetString(FortyFiveUpIcon, "↗")
	case "Flat":
		return ns.props.GetString(FlatIcon, "→")
	case "FortyFiveDown":
		return ns.props.GetString(FortyFiveDownIcon, "↘")
	case "SingleDown":
		return ns.props.GetString(SingleDownIcon, "↓")
	case "DoubleDown":
		return ns.props.GetString(DoubleDownIcon, "↓↓")
	default:
		return ""
	}
}

func (ns *nightscout) getResult() (*NightscoutData, error) {
	parseSingleElement := func(data []byte) (*NightscoutData, error) {
		var result []*NightscoutData
		err := json.Unmarshal(data, &result)
		if err != nil {
			return nil, err
		}
		if len(result) == 0 {
			return nil, errors.New("no elements in the array")
		}
		return result[0], nil
	}
	getCacheValue := func(key string) (*NightscoutData, error) {
		val, found := ns.env.Cache().Get(key)
		// we got something from the cache
		if found {
			if data, err := parseSingleElement([]byte(val)); err == nil {
				return data, nil
			}
		}
		return nil, errors.New("no data in cache")
	}

	url := ns.props.GetString(URL, "")
	httpTimeout := ns.props.GetInt(HTTPTimeout, DefaultHTTPTimeout)
	// natural and understood NS timeout is 5, anything else is unusual
	cacheTimeout := ns.props.GetInt(NSCacheTimeout, 5)

	if cacheTimeout > 0 {
		if data, err := getCacheValue(url); err == nil {
			return data, nil
		}
	}

	body, err := ns.env.HTTPRequest(url, httpTimeout)
	if err != nil {
		return nil, err
	}
	var arr []*NightscoutData
	err = json.Unmarshal(body, &arr)
	if err != nil {
		return nil, err
	}

	data, err := parseSingleElement(body)
	if err != nil {
		return nil, err
	}

	if cacheTimeout > 0 {
		// persist new sugars in cache
		ns.env.Cache().Set(url, string(body), cacheTimeout)
	}
	return data, nil
}

func (ns *nightscout) init(props properties.Properties, env environment.Environment) {
	ns.props = props
	ns.env = env
}
