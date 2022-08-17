package segments

import (
	"encoding/json"
	"oh-my-posh/environment"
	"oh-my-posh/properties"
	"oh-my-posh/template"
)

type Wakatime struct {
	props properties.Properties
	env   environment.Environment

	wtData
}

type wtTotals struct {
	Seconds float64 `json:"seconds"`
	Text    string  `json:"text"`
}

type wtData struct {
	CummulativeTotal wtTotals `json:"cummulative_total"`
	Start            string   `json:"start"`
	End              string   `json:"end"`
}

func (w *Wakatime) Template() string {
	return " {{ secondsRound .CummulativeTotal.Seconds }} "
}

func (w *Wakatime) Enabled() bool {
	err := w.setAPIData()
	return err == nil
}

func (w *Wakatime) setAPIData() error {
	url, err := w.getURL()
	if err != nil {
		return err
	}
	cacheTimeout := w.props.GetInt(properties.CacheTimeout, properties.DefaultCacheTimeout)
	if cacheTimeout > 0 {
		// check if data stored in cache
		if val, found := w.env.Cache().Get(url); found {
			err := json.Unmarshal([]byte(val), &w.wtData)
			if err != nil {
				return err
			}
			return nil
		}
	}

	httpTimeout := w.props.GetInt(properties.HTTPTimeout, properties.DefaultHTTPTimeout)

	body, err := w.env.HTTPRequest(url, nil, httpTimeout)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &w.wtData)
	if err != nil {
		return err
	}

	if cacheTimeout > 0 {
		w.env.Cache().Set(url, string(body), cacheTimeout)
	}
	return nil
}

func (w *Wakatime) getURL() (string, error) {
	url := w.props.GetString(URL, "")
	tmpl := &template.Text{
		Template: url,
		Context:  w,
		Env:      w.env,
	}
	return tmpl.Render()
}

func (w *Wakatime) Init(props properties.Properties, env environment.Environment) {
	w.props = props
	w.env = env
}
