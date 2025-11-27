package auth

import (
	"encoding/json"
	"fmt"
	httplib "net/http"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/http"
)

const (
	// GitHub Copilot's OAuth client ID - This is a public client ID used for device code flow
	CopilotClientID = "Iv1.b507a08c87ecfe98"
	CopilotScope    = "read:email"

	CopilotDeviceCodeURL  = "https://github.com/login/device/code"
	CopilotAccessTokenURL = "https://github.com/login/oauth/access_token"

	CopilotTokenKey = "copilot_token"
)

// DeviceCodeResponse represents the response from GitHub's device code endpoint.
type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

// AccessTokenResponse represents the response from GitHub's access token endpoint.
type AccessTokenResponse struct {
	AccessToken      string `json:"access_token"`
	TokenType        string `json:"token_type"`
	Scope            string `json:"scope"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func NewCopilot(env runtime.Environment) *CopilotAuth {
	return &CopilotAuth{
		model: model{
			env: env,
		},
	}
}

type CopilotAuth struct {
	deviceCodeExpiry time.Time
	verificationURI  string
	model
	lastState state
}

func (c *CopilotAuth) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case stateMsg:
		c.state = state(msg)
		if c.state == done {
			return c, tea.Quit
		}

		return c, nil

	default:
		s, cmd := c.spinner.Update(msg)
		c.spinner = &s
		return c, cmd
	}
}

func (c *CopilotAuth) Init() tea.Cmd {
	c.model.status = c.status
	cmd := c.model.Init()
	go c.Authenticate()
	return cmd
}

func (c *CopilotAuth) Authenticate() {
	setState(code)
	c.lastState = code

	deviceCode, err := c.requestDeviceCode()
	if err != nil {
		c.err = err
		setState(done)
		return
	}

	c.code = deviceCode.UserCode
	c.verificationURI = deviceCode.VerificationURI
	c.deviceCodeExpiry = time.Now().Add(time.Duration(deviceCode.ExpiresIn) * time.Second)

	setState(token)
	c.lastState = token

	interval := max(deviceCode.Interval, 5)

	token, err := c.pollForToken(deviceCode.DeviceCode, interval)
	if err != nil {
		c.err = err
		setState(done)
		return
	}

	if token == "" {
		c.err = fmt.Errorf("received empty token")
		setState(done)
		return
	}

	cache.Set(cache.Device, CopilotTokenKey, token, cache.TWOYEARS)

	setState(done)
}

func (c *CopilotAuth) requestDeviceCode() (*DeviceCodeResponse, error) {
	body := fmt.Sprintf("client_id=%s&scope=%s", CopilotClientID, CopilotScope)

	modifyRequest := func(request *httplib.Request) {
		request.Method = httplib.MethodPost
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		request.Header.Set("Accept", "application/json")
	}

	response, err := c.env.HTTPRequest(CopilotDeviceCodeURL, strings.NewReader(body), 30000, modifyRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to request device code: %w", err)
	}

	var result DeviceCodeResponse
	if err := json.Unmarshal(response, &result); err != nil {
		return nil, fmt.Errorf("failed to parse device code response: %w", err)
	}

	return &result, nil
}

func (c *CopilotAuth) pollForToken(deviceCode string, interval int) (string, error) {
	modifyRequest := func(request *httplib.Request) {
		request.Method = httplib.MethodPost
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		request.Header.Set("Accept", "application/json")
	}

	for {
		if time.Now().After(c.deviceCodeExpiry) {
			return "", fmt.Errorf("device code expired, please try again")
		}

		time.Sleep(time.Duration(interval) * time.Second)

		body := fmt.Sprintf("client_id=%s&device_code=%s&grant_type=urn:ietf:params:oauth:grant-type:device_code", CopilotClientID, deviceCode)
		response, err := c.env.HTTPRequest(CopilotAccessTokenURL, strings.NewReader(body), 30000, modifyRequest)
		if err != nil {
			// Log error but continue polling
			continue
		}

		var result AccessTokenResponse
		if err := json.Unmarshal(response, &result); err != nil {
			// Log error but continue polling
			continue
		}

		if result.AccessToken != "" {
			return result.AccessToken, nil
		}

		switch result.Error {
		case "authorization_pending":
			continue
		case "slow_down":
			interval += 5
			continue
		case "expired_token":
			return "", fmt.Errorf("device code expired, please try again")
		case "access_denied":
			return "", fmt.Errorf("access was denied by the user")
		default:
			if result.Error != "" {
				return "", fmt.Errorf("authentication error: %s - %s", result.Error, result.ErrorDescription)
			}
		}
	}
}

func (c *CopilotAuth) status(err error) string {
	if err == nil {
		return "Successfully authenticated with GitHub Copilot"
	}

	httpErr, ok := err.(*http.Error)
	if !ok {
		return err.Error()
	}

	return fmt.Sprintf("HTTP error %d: %s", httpErr.StatusCode, httpErr.Error())
}

func (c *CopilotAuth) View() string {
	var message string

	switch c.state {
	case code:
		message = fmt.Sprintf("%s Requesting device code from GitHub", c.spinner.View())
	case token:
		message = fmt.Sprintf("%s Please visit %s and enter code: %s", c.spinner.View(), c.verificationURI, c.code)
	case done:
		message = c.status(c.err)
	}

	return textStyle.Render(message)
}
