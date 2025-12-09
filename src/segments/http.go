package segments

import (
	"encoding/json"

	"net/http"

	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
)

type HTTP struct {
	Base

	Body map[string]any
}

const (
	METHOD options.Option = "method"
)

func (h *HTTP) Template() string {
	return " {{ .Body }} "
}

func (h *HTTP) Enabled() bool {
	url := h.options.String(URL, "")
	if url == "" {
		return false
	}

	method := h.options.String(METHOD, "GET")

	if resolved, err := template.Render(url, nil); err == nil {
		url = resolved
	}

	result, err := h.getResult(url, method)
	if err != nil {
		return false
	}

	h.Body = result
	return true
}

func (h *HTTP) getResult(url, method string) (map[string]any, error) {
	setMethod := func(request *http.Request) {
		request.Method = method
	}

	resultBody, err := h.env.HTTPRequest(url, nil, 10000, setMethod)
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
