package main

import "oh-my-posh/environment"

type ipify struct {
	props Properties
	env   environment.Environment
	IP    string
}

const (
	IpifyURL Property = "url"
)

func (i *ipify) template() string {
	return "{{ .IP }}"
}

func (i *ipify) enabled() bool {
	ip, err := i.getResult()
	if err != nil {
		return false
	}
	i.IP = ip

	return true
}

func (i *ipify) getResult() (string, error) {
	cacheTimeout := i.props.getInt(CacheTimeout, DefaultCacheTimeout)

	url := i.props.getString(IpifyURL, "https://api.ipify.org")

	if cacheTimeout > 0 {
		// check if data stored in cache
		val, found := i.env.Cache().Get(url)
		// we got something from te cache
		if found {
			return val, nil
		}
	}

	httpTimeout := i.props.getInt(HTTPTimeout, DefaultHTTPTimeout)

	body, err := i.env.HTTPRequest(url, httpTimeout)
	if err != nil {
		return "", err
	}

	// convert the body to a string
	response := string(body)

	if cacheTimeout > 0 {
		// persist public ip in cache
		i.env.Cache().Set(url, response, cacheTimeout)
	}
	return response, nil
}

func (i *ipify) init(props Properties, env environment.Environment) {
	i.props = props
	i.env = env
}
