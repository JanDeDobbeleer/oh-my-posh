package segments

import (
	"encoding/json"

	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
)

type Wakatime struct {
	Base

	WtData
}

type wtTotals struct {
	Text    string  `json:"text"`
	Seconds float64 `json:"seconds"`
}

type WtData struct {
	Start           string   `json:"start"`
	End             string   `json:"end"`
	CumulativeTotal wtTotals `json:"cumulative_total"`
}

func (w *Wakatime) Template() string {
	return " {{ secondsRound .CumulativeTotal.Seconds }} "
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

	httpTimeout := w.options.Int(options.HTTPTimeout, options.DefaultHTTPTimeout)

	body, err := w.env.HTTPRequest(url, nil, httpTimeout)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, &w.WtData)
	if err != nil {
		return err
	}

	return nil
}

func (w *Wakatime) getURL() (string, error) {
	url := w.options.String(URL, "")
	return template.Render(url, w)
}
