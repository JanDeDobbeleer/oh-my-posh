package http

import (
	"encoding/json"
	"fmt"
	"io"
	httplib "net/http"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
)

const (
	Timeout             = "timeout"
	InvalidRefreshToken = "invalid refresh token"
	TokenRefreshFailed  = "token refresh error"
	DefaultRefreshToken = "111111111111111111111111111111"
)

type tokenExchange struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type OAuthError struct {
	message string
}

func (a *OAuthError) Error() string {
	return a.message
}

type OAuthRequest struct {
	Request

	AccessTokenKey  string
	RefreshTokenKey string
	SegmentName     string

	RefreshToken string
	AccessToken  string
}

func (o *OAuthRequest) getAccessToken() (string, error) {
	// get directly from cache
	if acccessToken, OK := o.Env.Cache().Get(o.AccessTokenKey); OK && len(acccessToken) != 0 {
		return acccessToken, nil
	}

	// use cached refresh token to get new access token
	if refreshToken, OK := o.Env.Cache().Get(o.RefreshTokenKey); OK && len(refreshToken) != 0 {
		if acccessToken, err := o.refreshToken(refreshToken); err == nil {
			return acccessToken, nil
		}
	}

	// use initial refresh token from property
	// refreshToken := o.props.GetString(properties.RefreshToken, "")
	// ignore an empty or default refresh token
	if len(o.RefreshToken) == 0 || o.RefreshToken == DefaultRefreshToken {
		return "", &OAuthError{
			message: InvalidRefreshToken,
		}
	}

	// no need to let the user provide access token, we'll always verify the refresh token
	acccessToken, err := o.refreshToken(o.RefreshToken)
	return acccessToken, err
}

func (o *OAuthRequest) refreshToken(refreshToken string) (string, error) {
	// httpTimeout := o.props.GetInt(properties.HTTPTimeout, properties.DefaultHTTPTimeout)
	if o.HTTPTimeout == 0 {
		o.HTTPTimeout = 20
	}

	url := fmt.Sprintf("https://ohmyposh.dev/api/refresh?segment=%s&token=%s", o.SegmentName, refreshToken)
	body, err := o.Env.HTTPRequest(url, nil, o.HTTPTimeout)
	if err != nil {
		return "", &OAuthError{
			// This might happen if /api was asleep. Assume the user will just retry
			message: Timeout,
		}
	}

	tokens := &tokenExchange{}
	err = json.Unmarshal(body, &tokens)
	if err != nil {
		return "", &OAuthError{
			message: TokenRefreshFailed,
		}
	}

	// add tokens to cache
	o.Env.Cache().Set(o.AccessTokenKey, tokens.AccessToken, cache.ToDuration(tokens.ExpiresIn))
	o.Env.Cache().Set(o.RefreshTokenKey, tokens.RefreshToken, "2years")
	return tokens.AccessToken, nil
}

func OauthResult[a any](o *OAuthRequest, url string, body io.Reader, requestModifiers ...RequestModifier) (a, error) {
	accessToken, err := o.getAccessToken()
	if err != nil {
		var data a
		return data, err
	}

	// add token to header for authentication
	addAuthHeader := func(request *httplib.Request) {
		request.Header.Add("Authorization", "Bearer "+accessToken)
	}

	if requestModifiers == nil {
		requestModifiers = []RequestModifier{}
	}

	requestModifiers = append(requestModifiers, addAuthHeader)

	return Do[a](&o.Request, url, body, requestModifiers...)
}
