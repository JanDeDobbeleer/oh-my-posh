package main

import (
	"encoding/json"
	"errors"
	"time"
)

// segment struct, makes templating easier
type nightscout struct {
	props *properties
	env   environmentInfo

	NightscoutData
	TrendIcon string
}

const (
	// Your complete Nightscout URL and APIKey like this
	URL Property = "url"

	DoubleUpIcon      Property = "doubleup_icon"
	SingleUpIcon      Property = "singleup_icon"
	FortyFiveUpIcon   Property = "fortyfiveup_icon"
	FlatIcon          Property = "flat_icon"
	FortyFiveDownIcon Property = "fortyfivedown_icon"
	SingleDownIcon    Property = "singledown_icon"
	DoubleDownIcon    Property = "doubledown_icon"

	NSCacheTimeout Property = "cache_timeout"
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
		return ns.props.getString(DoubleUpIcon, "↑↑")
	case "SingleUp":
		return ns.props.getString(SingleUpIcon, "↑")
	case "FortyFiveUp":
		return ns.props.getString(FortyFiveUpIcon, "↗")
	case "Flat":
		return ns.props.getString(FlatIcon, "→")
	case "FortyFiveDown":
		return ns.props.getString(FortyFiveDownIcon, "↘")
	case "SingleDown":
		return ns.props.getString(SingleDownIcon, "↓")
	case "DoubleDown":
		return ns.props.getString(DoubleDownIcon, "↓↓")
	default:
		return ""
	}
}

func (ns *nightscout) string() string {
	segmentTemplate := ns.props.getString(SegmentTemplate, "{{.Sgv}}")
	template := &textTemplate{
		Template: segmentTemplate,
		Context:  ns,
		Env:      ns.env,
	}
	text, err := template.render()
	if err != nil {
		return err.Error()
	}

	return text
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
		val, found := ns.env.cache().get(key)
		// we got something from the cache
		if found {
			if data, err := parseSingleElement([]byte(val)); err == nil {
				return data, nil
			}
		}
		return nil, errors.New("no data in cache")
	}

	url := ns.props.getString(URL, "")
	httpTimeout := ns.props.getInt(HTTPTimeout, DefaultHTTPTimeout)
	// natural and understood NS timeout is 5, anything else is unusual
	cacheTimeout := ns.props.getInt(NSCacheTimeout, 5)

	if cacheTimeout > 0 {
		if data, err := getCacheValue(url); err == nil {
			return data, nil
		}
	}

	body, err := ns.env.doGet(url, httpTimeout)
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
		ns.env.cache().set(url, string(body), cacheTimeout)
	}
	return data, nil
}

func (ns *nightscout) init(props *properties, env environmentInfo) {
	ns.props = props
	ns.env = env
}
