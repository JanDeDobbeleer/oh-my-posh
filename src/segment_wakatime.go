package main

import (
	"encoding/json"
	"fmt"
	"math"
)

type wakatime struct {
	props properties
	env   environmentInfo

	Hours        int
	Minutes      int
	MinutesTotal int
	URL          string
	Response     *wtDataResponse
}

const (
	// CacheKeyResponse key used when caching the response
	WTCacheKeyResponse string = "wt_response"
	// WTCacheKeyURL key used when caching the url
	WTCacheKeyURL string = "wt_url"
)

type wtTotals struct {
	Decimal string  `json:"decimal"`
	Digital string  `json:"digital"`
	Seconds float64 `json:"seconds"`
	Text    string  `json:"text"`
}

type wtDataResponse struct {
	CummulativeTotal wtTotals `json:"cummulative_total"`
	Start            string   `json:"start"`
	End              string   `json:"end"`
}

func (w *wakatime) enabled() bool {
	err := w.setStatus()
	return err == nil
}

func (w *wakatime) string() string {
	segmentTemplate := w.props.getString(SegmentTemplate, "{{ if gt .Hours 0 }}{{.Hours}}h {{ end }}{{.Minutes}}m")
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

func (w *wakatime) getResult() (*wtDataResponse, error) {
	cacheTimeout := w.props.getInt(CacheTimeout, DefaultCacheTimeout)
	w.Response = &wtDataResponse{}
	if cacheTimeout > 0 {
		// check if data stored in cache
		if val, found := w.env.cache().get(WTCacheKeyResponse); found {
			err := json.Unmarshal([]byte(val), w.Response)
			if err != nil {
				return nil, err
			}
			w.URL, _ = w.env.cache().get(WTCacheKeyURL)
			return w.Response, nil
		}
	}

	apikey := w.props.getString(APIKey, ".")
	httpTimeout := w.props.getInt(HTTPTimeout, DefaultHTTPTimeout)
	w.URL = fmt.Sprintf("https://wakatime.com/api/v1/users/current/summaries?start=today&end=today&api_key=%s", apikey)

	body, err := w.env.doGet(w.URL, httpTimeout)
	if err != nil {
		return new(wtDataResponse), err
	}
	err = json.Unmarshal(body, w.Response)
	if err != nil {
		return new(wtDataResponse), err
	}

	if cacheTimeout > 0 {
		// persist data in cache
		w.env.cache().set(WTCacheKeyURL, w.URL, cacheTimeout)
		w.env.cache().set(WTCacheKeyResponse, string(body), cacheTimeout)
	}
	return w.Response, nil
}

func (w *wakatime) setStatus() error {
	q, err := w.getResult()
	if err != nil {
		return err
	}

	w.Hours = int(math.Floor(q.CummulativeTotal.Seconds / 3600))
	w.MinutesTotal = int(math.Floor(q.CummulativeTotal.Seconds / 60))
	w.Minutes = w.MinutesTotal % 60
	return nil
}

func (w *wakatime) init(props properties, env environmentInfo) {
	w.props = props
	w.env = env
}
