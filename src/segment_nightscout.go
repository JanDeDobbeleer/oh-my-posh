package main

import (
	"encoding/json"
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
	Sgv       int64  `json:"sgv"`
	Direction string `json:"direction"`
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
	url := ns.props.getString(URL, "")
	// natural and understood NS timeout is 5, anything else is unusual
	cacheTimeout := ns.props.getInt(NSCacheTimeout, 5)
	response := &NightscoutData{}
	if cacheTimeout > 0 {
		// check if data stored in cache
		val, found := ns.env.cache().get(url)
		// we got something from the cache
		if found {
			err := json.Unmarshal([]byte(val), response)
			if err != nil {
				return nil, err
			}
			return response, nil
		}
	}

	httpTimeout := ns.props.getInt(HTTPTimeout, DefaultHTTPTimeout)

	body, err := ns.env.doGet(url, httpTimeout)
	if err != nil {
		return &NightscoutData{}, err
	}
	var arr []*NightscoutData
	err = json.Unmarshal(body, &arr)
	if err != nil {
		return &NightscoutData{}, err
	}

	firstelement := arr[0]
	firstData, err := json.Marshal(firstelement)
	if err != nil {
		return &NightscoutData{}, err
	}

	if cacheTimeout > 0 {
		// persist new sugars in cache
		ns.env.cache().set(url, string(firstData), cacheTimeout)
	}
	return firstelement, nil
}

func (ns *nightscout) init(props *properties, env environmentInfo) {
	ns.props = props
	ns.env = env
}
