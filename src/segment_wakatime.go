package main

import (
	"encoding/json"
	"math"
	neturl "net/url"
	"strconv"
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
	// Start date filter
	Start Property = "start"
	// End date filter
	End Property = "end"
	// Project filter
	Project Property = "project"
	// Branches filter
	Branches Property = "branches"
	// Timout to merge heartbeats into durations
	Timeout Property = "timeout"
	// Only writes string
	WritesOnly Property = "writes_only"
	// Timezone for start and end date filters
	Timezone Property = "timezone"
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

	httpTimeout := w.props.getInt(HTTPTimeout, DefaultHTTPTimeout)
	// Build API URL
	w.setAPIURL()

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

func (w *wakatime) setAPIURL() {
	url := neturl.URL{
		Scheme: "https",
		Host:   "wakatime.com",
		Path:   "/api/v1/users/current/summaries",
	}

	w.addURLStringParam(&url, APIKey, ".", "api_key")
	w.addURLStringParam(&url, Start, "today", "start")
	w.addURLStringParam(&url, End, "today", "end")
	w.addURLStringParam(&url, Timezone, "", "timezone")
	w.addURLStringParam(&url, Project, "", "project")
	w.addURLStringParam(&url, Branches, "", "branches")
	w.addURLIntParam(&url, Timeout, 0, "timeout")
	w.addURLBoolParam(&url, WritesOnly, false, "writes_only")

	w.URL = url.String()
}

func (w *wakatime) addURLStringParam(url *neturl.URL, property Property, defValue, paramName string) {
	value := w.props.getString(property, defValue)
	if len(value) > 0 {
		q := url.Query()
		q.Set(paramName, value)
		url.RawQuery = q.Encode()
	}
}

func (w *wakatime) addURLIntParam(url *neturl.URL, property Property, defValue int, paramName string) {
	value := w.props.getInt(property, defValue)
	if value > 0 {
		q := url.Query()
		q.Set(paramName, strconv.Itoa(value))
		url.RawQuery = q.Encode()
	}
}

func (w *wakatime) addURLBoolParam(url *neturl.URL, property Property, defValue bool, paramName string) {
	value := w.props.getBool(property, defValue)
	if value {
		q := url.Query()
		q.Set(paramName, "true")
		url.RawQuery = q.Encode()
	}
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
