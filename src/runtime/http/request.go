package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
)

type RequestModifier func(request *http.Request)

type Request struct {
	Env          Environment
	CacheTimeout int
	HTTPTimeout  int
}

type Environment interface {
	HTTPRequest(url string, body io.Reader, timeout int, requestModifiers ...RequestModifier) ([]byte, error)
	Cache() cache.Cache
}

func Do[a any](r *Request, url string, requestModifiers ...RequestModifier) (a, error) {
	if data, err := getCacheValue[a](r, url); err == nil {
		return data, nil
	}
	return do[a](r, url, nil, requestModifiers...)
}

func getCacheValue[a any](r *Request, key string) (a, error) {
	var data a
	cacheTimeout := r.CacheTimeout // r.props.GetInt(properties.CacheTimeout, 30)
	if cacheTimeout <= 0 {
		return data, errors.New("no cache needed")
	}

	if val, found := r.Env.Cache().Get(key); found {
		err := json.Unmarshal([]byte(val), &data)
		if err != nil {
			log.Error(err)
			return data, err
		}

		return data, nil
	}

	err := errors.New("no data in cache")
	log.Error(err)

	return data, err
}

func do[a any](r *Request, url string, body io.Reader, requestModifiers ...RequestModifier) (a, error) {
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

	cacheTimeout := r.CacheTimeout // r.props.GetInt(properties.CacheTimeout, 30)
	if cacheTimeout > 0 {
		r.Env.Cache().Set(url, string(responseBody), cacheTimeout)
	}

	return data, nil
}
