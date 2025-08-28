package segments

import (
	"encoding/json"

	"net/http"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
)

type HTTP struct {
	base

	Body map[string]any
}

const (
	METHOD  properties.Property = "method"
	TIMEOUT properties.Property = "timeout"
)

func (h *HTTP) Template() string {
	return " {{ .Body }} "
}

func (h *HTTP) Enabled() bool {
	url := h.props.GetString(URL, "")
	if url == "" {
		return false
	}

	method := h.props.GetString(METHOD, "GET")

	timeout := h.props.GetInt(TIMEOUT, 10000)

	if resolved, err := template.Render(url, nil); err == nil {
		url = resolved
	}

	result, err := h.getResult(url, method, timeout)
	if err != nil {
		return false
	}

	h.Body = result
	return true
}

func (h *HTTP) getResult(url, method string, timeout int) (map[string]any, error) {
	setMethod := func(request *http.Request) {
		request.Method = method
	}

	resultBody, err := h.env.HTTPRequest(url, nil, timeout, setMethod)
	if err != nil {
		return nil, err
	}

	var result map[string]any
	err = json.Unmarshal(resultBody, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
