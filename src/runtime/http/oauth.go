//revive:disable:var-naming // package intentionally mirrors standard name for compatibility across runtime
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
	AccessTokenKey  string
	RefreshTokenKey string
	SegmentName     string
	RefreshToken    string
	AccessToken     string
	Request
}

func (o *OAuthRequest) getAccessToken() (string, error) {
	// get directly from cache
	if accessToken, OK := cache.Device.Get[string](o.AccessTokenKey); OK && len(accessToken) != 0 {
		return accessToken, nil
	}

	// use cached refresh token to get new access token
	if refreshToken, OK := cache.Device.Get[string](o.RefreshTokenKey); OK && len(refreshToken) != 0 {
		if accessToken, err := o.refreshToken(refreshToken); err == nil {
			return accessToken, nil
		}
	}

	// use initial refresh token from property
	// refreshToken := o.props.GetString(options.RefreshToken, "")
	// ignore an empty or default refresh token
	if o.RefreshToken == "" || o.RefreshToken == DefaultRefreshToken {
		return "", &OAuthError{
			message: InvalidRefreshToken,
		}
	}

	// no need to let the user provide access token, we'll always verify the refresh token
	accessToken, err := o.refreshToken(o.RefreshToken)
	return accessToken, err
}

func (o *OAuthRequest) refreshToken(refreshToken string) (string, error) {
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
	cache.Device.Set(o.AccessTokenKey, tokens.AccessToken, cache.ToDuration(tokens.ExpiresIn))
	cache.Device.Set(o.RefreshTokenKey, tokens.RefreshToken, cache.TWOYEARS)
	return tokens.AccessToken, nil
}

func (o *OAuthRequest) Result[T any](url string, body io.Reader, requestModifiers ...RequestModifier) (T, error) {
	accessToken, err := o.getAccessToken()
	if err != nil {
		var data T
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

	return o.Do[T](url, body, requestModifiers...)
}
