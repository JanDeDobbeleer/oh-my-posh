package segments

import (
	"net"

	"github.com/jandedobbeleer/oh-my-posh/src/http"
	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
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
	return http.Do[*ipData](&i.Request, url)
}

type IPify struct {
	IP string

	api IPAPI
}

const (
	IpifyURL properties.Property = "url"

	OFFLINE = "OFFLINE"
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
	data, err := i.api.Get()
	if dnsErr, OK := err.(*net.DNSError); OK && dnsErr.IsNotFound {
		return OFFLINE, nil
	}
	if err != nil {
		return "", err
	}
	return data.IP, err
}

func (i *IPify) Init(props properties.Properties, env platform.Environment) {
	request := &http.Request{}
	request.Init(env, props)

	i.api = &ipAPI{
		Request: *request,
	}
}
