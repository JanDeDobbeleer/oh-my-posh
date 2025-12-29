package segments

import (
	"net"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/http"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
)

type ipData struct {
	IP string `json:"ip"`
}

type IPAPI interface {
	Get() (*ipData, error)
}

type ipAPI struct {
	http.Request
}

func (i *ipAPI) Get() (*ipData, error) {
	url := "https://api.ipify.org?format=json"
	return http.Do[*ipData](&i.Request, url, nil)
}

type IPify struct {
	Base

	api IPAPI
	IP  string
}

const (
	OFFLINE = "OFFLINE"
)

func (i *IPify) Template() string {
	return " {{ .IP }} "
}

func (i *IPify) Enabled() bool {
	const key = "IP"

	if ip, ok := cache.Get[string](cache.Device, key); ok {
		i.IP = ip
		return true
	}

	i.initAPI()

	ip, err := i.getResult()
	if err != nil {
		return false
	}

	i.IP = ip

	duration := i.options.String(options.CacheDuration, string(cache.ONEDAY))
	cache.Set(cache.Device, key, i.IP, cache.Duration(duration))

	return true
}

func (i *IPify) getResult() (string, error) {
	data, err := i.api.Get()
	if dnsErr, OK := err.(*net.DNSError); OK && dnsErr.IsNotFound {
		return OFFLINE, nil
	}

	if err != nil {
		return "", err
	}

	return data.IP, err
}

func (i *IPify) initAPI() {
	if i.api != nil {
		return
	}

	request := &http.Request{
		Env:         i.env,
		HTTPTimeout: i.options.Int(options.HTTPTimeout, options.DefaultHTTPTimeout),
	}

	i.api = &ipAPI{
		Request: *request,
	}
}
