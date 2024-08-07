package segments

import (
	"encoding/json"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
)

type Wakatime struct {
	props properties.Properties
	env   runtime.Environment

	wtData
}

type wtTotals struct {
	Text    string  `json:"text"`
	Seconds float64 `json:"seconds"`
}

type wtData struct {
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

	httpTimeout := w.props.GetInt(properties.HTTPTimeout, properties.DefaultHTTPTimeout)

	body, err := w.env.HTTPRequest(url, nil, httpTimeout)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, &w.wtData)
	if err != nil {
		return err
	}

	return nil
}

func (w *Wakatime) getURL() (string, error) {
	url := w.props.GetString(URL, "")
	tmpl := &template.Text{
		Template: url,
		Context:  w,
	}
	return tmpl.Render()
}

func (w *Wakatime) Init(props properties.Properties, env runtime.Environment) {
	w.props = props
	w.env = env
}
