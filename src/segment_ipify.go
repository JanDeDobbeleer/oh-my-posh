package main

type ipify struct {
	props Properties
	env   environmentInfo
	IP    string
}

const (
	IpifyURL Property = "url"
)

func (i *ipify) enabled() bool {
	ip, err := i.getResult()
	if err != nil {
		return false
	}
	i.IP = ip

	return true
}

func (i *ipify) string() string {
	segmentTemplate := i.props.getString(SegmentTemplate, "{{.IP}}")
	template := &textTemplate{
		Template: segmentTemplate,
		Context:  i,
		Env:      i.env,
	}
	text, err := template.render()
	if err != nil {
		return err.Error()
	}

	return text
}

func (i *ipify) getResult() (string, error) {
	cacheTimeout := i.props.getInt(CacheTimeout, DefaultCacheTimeout)

	url := i.props.getString(IpifyURL, "https://api.ipify.org")

	if cacheTimeout > 0 {
		// check if data stored in cache
		val, found := i.env.cache().get(url)
		// we got something from te cache
		if found {
			return val, nil
		}
	}

	httpTimeout := i.props.getInt(HTTPTimeout, DefaultHTTPTimeout)

	body, err := i.env.doGet(url, httpTimeout)
	if err != nil {
		return "", err
	}

	// convert the body to a string
	response := string(body)

	if cacheTimeout > 0 {
		// persist public ip in cache
		i.env.cache().set(url, response, cacheTimeout)
	}
	return response, nil
}

func (i *ipify) init(props Properties, env environmentInfo) {
	i.props = props
	i.env = env
}
