package segments

import (
	"encoding/json"
	"errors"
	http2 "net/http"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
)

// segment struct, makes templating easier
type Nightscout struct {
	Base

	TrendIcon string
	NightscoutData
}

const (
	// Your complete Nightscout URL and APIKey like this
	URL     options.Option = "url"
	Headers options.Option = "headers"

	DoubleUpIcon      options.Option = "doubleup_icon"
	SingleUpIcon      options.Option = "singleup_icon"
	FortyFiveUpIcon   options.Option = "fortyfiveup_icon"
	FlatIcon          options.Option = "flat_icon"
	FortyFiveDownIcon options.Option = "fortyfivedown_icon"
	SingleDownIcon    options.Option = "singledown_icon"
	DoubleDownIcon    options.Option = "doubledown_icon"
)

// NightscoutData struct contains the API data
type NightscoutData struct {
	DateString time.Time `json:"dateString"`
	SysTime    time.Time `json:"sysTime"`
	ID         string    `json:"_id"`
	Direction  string    `json:"direction"`
	Device     string    `json:"device"`
	Type       string    `json:"type"`
	Sgv        int       `json:"sgv"`
	Date       int64     `json:"date"`
	Trend      int       `json:"trend"`
	UtcOffset  int       `json:"utcOffset"`
	Mills      int64     `json:"mills"`
}

func (ns *Nightscout) Template() string {
	return " {{ .Sgv }} "
}

func (ns *Nightscout) Enabled() bool {
	data, err := ns.getResult()
	if err != nil {
		return false
	}
	ns.NightscoutData = *data
	ns.TrendIcon = ns.getTrendIcon()

	return true
}

func (ns *Nightscout) getTrendIcon() string {
	switch ns.Direction {
	case "DoubleUp":
		return ns.options.String(DoubleUpIcon, "↑↑")
	case "SingleUp":
		return ns.options.String(SingleUpIcon, "↑")
	case "FortyFiveUp":
		return ns.options.String(FortyFiveUpIcon, "↗")
	case "Flat":
		return ns.options.String(FlatIcon, "→")
	case "FortyFiveDown":
		return ns.options.String(FortyFiveDownIcon, "↘")
	case "SingleDown":
		return ns.options.String(SingleDownIcon, "↓")
	case "DoubleDown":
		return ns.options.String(DoubleDownIcon, "↓↓")
	default:
		return ""
	}
}

func (ns *Nightscout) getResult() (*NightscoutData, error) {
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

	url := ns.options.String(URL, "")
	httpTimeout := ns.options.Int(options.HTTPTimeout, options.DefaultHTTPTimeout)

	headers := ns.options.KeyValueMap(Headers, map[string]string{})
	modifiers := func(request *http2.Request) {
		for key, value := range headers {
			request.Header.Add(key, value)
		}
	}

	body, err := ns.env.HTTPRequest(url, nil, httpTimeout, modifiers)
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

	return data, nil
}
