package segments

import (
	"net"

	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/http"
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
	api IPAPI
	IP  string
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

func (i *IPify) Init(props properties.Properties, env runtime.Environment) {
	request := &http.Request{
		Env:         env,
		HTTPTimeout: props.GetInt(properties.HTTPTimeout, properties.DefaultHTTPTimeout),
	}

	i.api = &ipAPI{
		Request: *request,
	}
}
