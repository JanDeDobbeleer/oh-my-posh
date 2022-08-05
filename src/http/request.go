package http

import (
	"encoding/json"
	"errors"
	"io"
	"oh-my-posh/environment"
	"oh-my-posh/properties"
)

type Request struct {
	props properties.Properties
	env   environment.Environment
}

func (r *Request) Init(env environment.Environment, props properties.Properties) {
	r.env = env
	r.props = props
}

func Do[a any](r *Request, url string, requestModifiers ...environment.HTTPRequestModifier) (a, error) {
	return do[a](r, url, nil, nil, requestModifiers...)
}

func do[a any](r *Request, url string, body io.Reader, preRequestFunc func() error, requestModifiers ...environment.HTTPRequestModifier) (a, error) {
	var data a

	getCacheValue := func(key string) (a, error) {
		if val, found := r.env.Cache().Get(key); found {
			err := json.Unmarshal([]byte(val), &data)
			if err != nil {
				r.env.Log(environment.Error, "OAuth", err.Error())
				return data, err
			}
			return data, nil
		}
		err := errors.New("no data in cache")
		r.env.Log(environment.Error, "OAuth", err.Error())
		return data, err
	}

	httpTimeout := r.props.GetInt(properties.HTTPTimeout, properties.DefaultHTTPTimeout)

	// No need to check more than every 30 minutes by default
	cacheTimeout := r.props.GetInt(properties.CacheTimeout, 30)
	if cacheTimeout > 0 {
		if data, err := getCacheValue(url); err == nil {
			return data, nil
		}
	}

	if preRequestFunc != nil {
		if err := preRequestFunc(); err != nil {
			return data, err
		}
	}

	responseBody, err := r.env.HTTPRequest(url, body, httpTimeout, requestModifiers...)
	if err != nil {
		r.env.Log(environment.Error, "OAuth", err.Error())
		return data, err
	}

	err = json.Unmarshal(responseBody, &data)
	if err != nil {
		r.env.Log(environment.Error, "OAuth", err.Error())
		return data, err
	}

	if cacheTimeout > 0 {
		r.env.Cache().Set(url, string(responseBody), cacheTimeout)
	}

	return data, nil
}
