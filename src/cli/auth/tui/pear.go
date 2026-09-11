package tui

import (
	"errors"
	"net"
	httplib "net/http"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/cli/auth"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/http"
)

const pearAuthURL = auth.PEARBASEURL + "/auth/ohmyposh"

func NewPear(env runtime.Environment) *Pear {
	flow := &Pear{
		env: env,
	}

	flow.model.status = flow.status

	return flow
}

type Pear struct {
	model
}

func (p *Pear) Authenticate() {
	type tokenResponse struct {
		AccessToken string `json:"accessToken"`
	}

	result, err := ytmdaRequest[tokenResponse](httplib.MethodPost, pearAuthURL, "", p.env)
	if err != nil {
		p.err = err
		return
	}

	if result.AccessToken == "" {
		p.err = errors.New("received empty access token")
		return
	}

	cache.Device.Set(auth.PEARTOKEN, result.AccessToken, cache.INFINITE)
}

func (p *Pear) message() string {
	return "Requesting access token from Pear Desktop, please allow access in the pop-up window"
}

func (p *Pear) status(err error) string {
	if err == nil {
		return "Successfully authenticated with Pear Desktop"
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return "There was a timeout while trying to connect to the Pear Desktop API Server. Please try again"
	}

	httpErr, ok := err.(*http.Error)
	if !ok {
		// if the error is not an http.Error, the service isn't running
		return "Pear Desktop is not running, please enable the API Server plugin"
	}

	if httpErr.StatusCode == httplib.StatusForbidden {
		return "Access denied, please press Allow in the Pear Desktop pop-up window"
	}

	return err.Error()
}
