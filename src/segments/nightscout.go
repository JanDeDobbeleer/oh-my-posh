package segments

import (
	"encoding/json"
	"errors"
	http2 "net/http"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
)

// segment struct, makes templating easier
type Nightscout struct {
	props     properties.Properties
	env       runtime.Environment
	TrendIcon string
	NightscoutData
}

const (
	// Your complete Nightscout URL and APIKey like this
	URL     properties.Property = "url"
	Headers properties.Property = "headers"

	DoubleUpIcon      properties.Property = "doubleup_icon"
	SingleUpIcon      properties.Property = "singleup_icon"
	FortyFiveUpIcon   properties.Property = "fortyfiveup_icon"
	FlatIcon          properties.Property = "flat_icon"
	FortyFiveDownIcon properties.Property = "fortyfivedown_icon"
	SingleDownIcon    properties.Property = "singledown_icon"
	DoubleDownIcon    properties.Property = "doubledown_icon"
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

	url := ns.props.GetString(URL, "")
	httpTimeout := ns.props.GetInt(properties.HTTPTimeout, properties.DefaultHTTPTimeout)

	headers := ns.props.GetKeyValueMap(Headers, map[string]string{})
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

func (ns *Nightscout) Init(props properties.Properties, env runtime.Environment) {
	ns.props = props
	ns.env = env
}
