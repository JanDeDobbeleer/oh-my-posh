package segments

import (
	"encoding/json"

	"net/http"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
)

type HTTP struct {
	base

	URL    string
	Result map[string]interface{}
}

const (
	HTTPURL properties.Property = "url"
	METHOD  properties.Property = "method"
)

func (cs *HTTP) Template() string {
	return " {{ .Result }} "
}

func (cs *HTTP) Enabled() bool {
	url := cs.props.GetString(HTTPURL, "")
	if len(url) == 0 {
		return false
	}
	method := cs.props.GetString(METHOD, "GET")

	result, err := cs.getResult(url, method)
	if err != nil {
		return false
	}

	cs.Result = result
	return true
}

func (cs *HTTP) getResult(url, method string) (map[string]interface{}, error) {
	setMethod := func(request *http.Request) {
		request.Method = method
	}

	resultBody, err := cs.env.HTTPRequest(url, nil, 10000, setMethod)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	err = json.Unmarshal(resultBody, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
