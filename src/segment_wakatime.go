package main

import (
	"encoding/json"
	"oh-my-posh/environment"
	"oh-my-posh/properties"
)

type wakatime struct {
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

func (w *wakatime) template() string {
	return "{{ secondsRound .CummulativeTotal.Seconds }}"
}

func (w *wakatime) enabled() bool {
	err := w.setAPIData()
	return err == nil
}

func (w *wakatime) setAPIData() error {
	url := w.props.GetString(URL, "")
	cacheTimeout := w.props.GetInt(CacheTimeout, DefaultCacheTimeout)
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

	httpTimeout := w.props.GetInt(HTTPTimeout, DefaultHTTPTimeout)

	body, err := w.env.HTTPRequest(url, httpTimeout)
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

func (w *wakatime) init(props properties.Properties, env environment.Environment) {
	w.props = props
	w.env = env
}
