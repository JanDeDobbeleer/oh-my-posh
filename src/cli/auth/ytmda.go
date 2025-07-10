package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	httplib "net/http"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jandedobbeleer/oh-my-posh/src/build"
	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/http"
)

const (
	YTMDABASEURL = "http://localhost:9863/api/v1"
	YTMDATOKEN   = "ytmda_token"

	tokenURL = YTMDABASEURL + "/auth/request"
	codeURL  = YTMDABASEURL + "/auth/requestcode"
)

func NewYtmda(env runtime.Environment) *Ytmda {
	return &Ytmda{
		model: model{
			env: env,
		},
	}
}

type Ytmda struct {
	model
	lastState state
}

func (y *Ytmda) Init() tea.Cmd {
	y.model.status = y.status
	cmd := y.model.Init()
	go y.Authenticate()
	return cmd
}

func (y *Ytmda) Authenticate() {
	setState(code)
	y.lastState = code

	code, err := y.requestCode()
	if err != nil {
		y.err = err
		setState(done)
		return
	}

	y.code = code
	setState(token)
	y.lastState = token

	token, err := y.requestToken(code)
	if err != nil {
		y.err = err
		setState(done)
		return
	}

	if token == "" {
		y.err = fmt.Errorf("received empty token")
		setState(done)
		return
	}

	y.env.Cache().Set(YTMDATOKEN, token, cache.INFINITE)

	setState(done)
}

func (y *Ytmda) requestCode() (string, error) {
	body := fmt.Sprintf(`{"appId": "ohmyposh", "appName": "oh-my-posh", "appVersion": "%s"}`, strings.TrimPrefix(build.Version, "v"))

	type codeResponse struct {
		Code string `json:"code"`
	}

	result, err := ytmdaRequest[codeResponse](httplib.MethodPost, codeURL, body, y.env)

	return result.Code, err
}

func (y *Ytmda) requestToken(code string) (string, error) {
	body := fmt.Sprintf(`{"appId": "ohmyposh", "code": "%s"}`, code)

	type tokenResponse struct {
		Token string `json:"token"`
	}

	result, err := ytmdaRequest[tokenResponse](httplib.MethodPost, tokenURL, body, y.env)

	return result.Token, err
}

func ytmdaRequest[a any](method, url, body string, env runtime.Environment, requestModifiers ...http.RequestModifier) (a, error) {
	if requestModifiers == nil {
		requestModifiers = []http.RequestModifier{}
	}

	modifyRequest := func(request *httplib.Request) {
		request.Method = method
		request.Header.Set("Content-Type", "application/json")
	}

	requestModifiers = append(requestModifiers, modifyRequest)

	var result a

	response, err := env.HTTPRequest(url, strings.NewReader(body), 50000, requestModifiers...)
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(response, &result)
	return result, err
}

func (y *Ytmda) status(err error) string {
	// get the status code from the error if available
	if err == nil {
		return "Successfully authenticated with YouTube Music Desktop App"
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return "There was a timeout while trying to connect to the YouTube Music Desktop App Companion API. Please try again"
	}

	httpErr, ok := err.(*http.Error)
	if !ok {
		// if the error is not an http.Error, the service isn't running
		return "YouTube Music Desktop App is not running, please start the Companion API"
	}

	if httpErr.StatusCode != httplib.StatusForbidden {
		return err.Error()
	}

	if y.lastState == token {
		return "Failed to request token with code. Please press Allow in the pop-up window"
	}

	return "Please enable companion authorization in the YouTube Music Desktop App settings"
}
