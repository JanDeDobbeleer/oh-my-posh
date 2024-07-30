package http

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

type RequestModifier func(request *http.Request)

type Request struct {
	Env         Environment
	HTTPTimeout int
}

type Environment interface {
	HTTPRequest(url string, body io.Reader, timeout int, requestModifiers ...RequestModifier) ([]byte, error)
	Cache() cache.Cache
}

func Do[a any](r *Request, url string, body io.Reader, requestModifiers ...RequestModifier) (a, error) {
	var data a
	httpTimeout := r.HTTPTimeout // r.props.GetInt(properties.HTTPTimeout, properties.DefaultHTTPTimeout)

	responseBody, err := r.Env.HTTPRequest(url, body, httpTimeout, requestModifiers...)
	if err != nil {
		log.Error(err)
		return data, err
	}

	err = json.Unmarshal(responseBody, &data)
	if err != nil {
		log.Error(err)
		return data, err
	}

	return data, nil
}
