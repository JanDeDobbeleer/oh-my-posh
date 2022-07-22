package segments

import (
	"oh-my-posh/environment"
	"oh-my-posh/properties"
)

type IPify struct {
	props properties.Properties
	env   environment.Environment
	IP    string
}

const (
	IpifyURL properties.Property = "url"
)

func (i *IPify) Template() string {
	return " {{ .IP }} "
}

func (i *IPify) Enabled() bool {
	ip, err := i.getResult()
	if err != nil {
		return false
	}
	i.IP = ip

	return true
}

func (i *IPify) getResult() (string, error) {
	cacheTimeout := i.props.GetInt(properties.CacheTimeout, properties.DefaultCacheTimeout)

	url := i.props.GetString(IpifyURL, "https://api.ipify.org")

	if cacheTimeout > 0 {
		// check if data stored in cache
		val, found := i.env.Cache().Get(url)
		// we got something from te cache
		if found {
			return val, nil
		}
	}

	httpTimeout := i.props.GetInt(properties.HTTPTimeout, properties.DefaultHTTPTimeout)

	body, err := i.env.HTTPRequest(url, nil, httpTimeout)
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

func (i *IPify) Init(props properties.Properties, env environment.Environment) {
	i.props = props
	i.env = env
}
