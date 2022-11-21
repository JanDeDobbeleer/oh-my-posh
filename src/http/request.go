package http

import (
	"encoding/json"
	"errors"
	"io"
	"oh-my-posh/platform"
	"oh-my-posh/properties"
)

type Request struct {
	props properties.Properties
	env   platform.Environment
}

func (r *Request) Init(env platform.Environment, props properties.Properties) {
	r.env = env
	r.props = props
}

func Do[a any](r *Request, url string, requestModifiers ...platform.HTTPRequestModifier) (a, error) {
	if data, err := getCacheValue[a](r, url); err == nil {
		return data, nil
	}
	return do[a](r, url, nil, requestModifiers...)
}

func getCacheValue[a any](r *Request, key string) (a, error) {
	var data a
	cacheTimeout := r.props.GetInt(properties.CacheTimeout, 30)
	if cacheTimeout <= 0 {
		return data, errors.New("no cache needed")
	}
	if val, found := r.env.Cache().Get(key); found {
		err := json.Unmarshal([]byte(val), &data)
		if err != nil {
			r.env.Error("OAuth", err)
			return data, err
		}
		return data, nil
	}
	err := errors.New("no data in cache")
	r.env.Error("OAuth", err)
	return data, err
}

func do[a any](r *Request, url string, body io.Reader, requestModifiers ...platform.HTTPRequestModifier) (a, error) {
	var data a
	httpTimeout := r.props.GetInt(properties.HTTPTimeout, properties.DefaultHTTPTimeout)

	responseBody, err := r.env.HTTPRequest(url, body, httpTimeout, requestModifiers...)
	if err != nil {
		r.env.Error("OAuth", err)
		return data, err
	}

	err = json.Unmarshal(responseBody, &data)
	if err != nil {
		r.env.Error("OAuth", err)
		return data, err
	}

	cacheTimeout := r.props.GetInt(properties.CacheTimeout, 30)
	if cacheTimeout > 0 {
		r.env.Cache().Set(url, string(responseBody), cacheTimeout)
	}

	return data, nil
}
