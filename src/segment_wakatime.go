package main

import (
	"encoding/json"
)

type wakatime struct {
	props Properties
	env   Environment

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

func (w *wakatime) enabled() bool {
	err := w.setAPIData()
	return err == nil
}

func (w *wakatime) string() string {
	segmentTemplate := w.props.getString(SegmentTemplate, "{{ secondsRound .CummulativeTotal.Seconds }}")
	template := &textTemplate{
		Template: segmentTemplate,
		Context:  w,
		Env:      w.env,
	}
	text, err := template.render()
	if err != nil {
		return err.Error()
	}

	return text
}

func (w *wakatime) setAPIData() error {
	url := w.props.getString(URL, "")
	cacheTimeout := w.props.getInt(CacheTimeout, DefaultCacheTimeout)
	if cacheTimeout > 0 {
		// check if data stored in cache
		if val, found := w.env.cache().get(url); found {
			err := json.Unmarshal([]byte(val), &w.wtData)
			if err != nil {
				return err
			}
			return nil
		}
	}

	httpTimeout := w.props.getInt(HTTPTimeout, DefaultHTTPTimeout)

	body, err := w.env.doGet(url, httpTimeout)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, &w.wtData)
	if err != nil {
		return err
	}

	if cacheTimeout > 0 {
		w.env.cache().set(url, string(body), cacheTimeout)
	}
	return nil
}

func (w *wakatime) init(props Properties, env Environment) {
	w.props = props
	w.env = env
}
