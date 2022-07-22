package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"oh-my-posh/environment"
	"oh-my-posh/properties"
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

type OAuth struct {
	Props properties.Properties
	Env   environment.Environment

	AccessTokenKey  string
	RefreshTokenKey string
	SegmentName     string
}

func (o *OAuth) error(err error) {
	o.Env.Log(environment.Error, "OAuth", err.Error())
}

func (o *OAuth) getAccessToken() (string, error) {
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
	refreshToken := o.Props.GetString(properties.RefreshToken, "")
	// ignore an empty or default refresh token
	if len(refreshToken) == 0 || refreshToken == DefaultRefreshToken {
		return "", &OAuthError{
			message: InvalidRefreshToken,
		}
	}
	// no need to let the user provide access token, we'll always verify the refresh token
	acccessToken, err := o.refreshToken(refreshToken)
	return acccessToken, err
}

func (o *OAuth) refreshToken(refreshToken string) (string, error) {
	httpTimeout := o.Props.GetInt(properties.HTTPTimeout, properties.DefaultHTTPTimeout)
	url := fmt.Sprintf("https://ohmyposh.dev/api/refresh?segment=%s&token=%s", o.SegmentName, refreshToken)
	body, err := o.Env.HTTPRequest(url, nil, httpTimeout)
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
	o.Env.Cache().Set(o.AccessTokenKey, tokens.AccessToken, tokens.ExpiresIn/60)
	o.Env.Cache().Set(o.RefreshTokenKey, tokens.RefreshToken, 2*525960) // it should never expire unless revoked, default to 2 year
	return tokens.AccessToken, nil
}

func OauthResult[a any](o *OAuth, url string, body io.Reader, requestModifiers ...environment.HTTPRequestModifier) (a, error) {
	var data a

	getCacheValue := func(key string) (a, error) {
		if val, found := o.Env.Cache().Get(key); found {
			err := json.Unmarshal([]byte(val), &data)
			if err != nil {
				o.error(err)
				return data, err
			}
			return data, nil
		}
		err := errors.New("no data in cache")
		o.error(err)
		return data, err
	}

	httpTimeout := o.Props.GetInt(properties.HTTPTimeout, properties.DefaultHTTPTimeout)

	// No need to check more than every 30 minutes by default
	cacheTimeout := o.Props.GetInt(properties.CacheTimeout, 30)
	if cacheTimeout > 0 {
		if data, err := getCacheValue(url); err == nil {
			return data, nil
		}
	}
	accessToken, err := o.getAccessToken()
	if err != nil {
		return data, err
	}

	// add token to header for authentication
	addAuthHeader := func(request *http.Request) {
		request.Header.Add("Authorization", "Bearer "+accessToken)
	}

	if requestModifiers == nil {
		requestModifiers = []environment.HTTPRequestModifier{}
	}

	requestModifiers = append(requestModifiers, addAuthHeader)

	responseBody, err := o.Env.HTTPRequest(url, body, httpTimeout, requestModifiers...)
	if err != nil {
		o.error(err)
		return data, err
	}

	err = json.Unmarshal(responseBody, &data)
	if err != nil {
		o.error(err)
		return data, err
	}

	if cacheTimeout > 0 {
		o.Env.Cache().Set(url, string(responseBody), cacheTimeout)
	}

	return data, nil
}
