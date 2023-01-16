package http

import (
	"fmt"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"

	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
)

type data struct {
	Hello string `json:"hello"`
}

func TestOauthResult(t *testing.T) {
	accessTokenKey := "test_access_token"
	refreshTokenKey := "test_refresh_token"
	tokenResponse := `{ "access_token":"NEW_ACCESSTOKEN","refresh_token":"NEW_REFRESHTOKEN", "expires_in":1234 }`
	jsonResponse := `{ "hello":"world" }`
	successData := &data{Hello: "world"}

	cases := []struct {
		Case string
		// tokens
		AccessToken   string
		RefreshToken  string
		TokenResponse string
		// API response
		JSONResponse string
		// Cache
		CacheJSONResponse     string
		CacheTimeout          int
		RefreshTokenFromCache bool
		AccessTokenFromCache  bool
		ResponseCacheMiss     bool
		// Errors
		Error error
		// Validations
		ExpectedErrorMessage string
		ExpectedData         *data
	}{
		{
			Case:                 "No initial tokens",
			ExpectedErrorMessage: InvalidRefreshToken,
		},
		{
			Case:          "Use config tokens",
			AccessToken:   "INITIAL_ACCESSTOKEN",
			RefreshToken:  "INITIAL_REFRESHTOKEN",
			TokenResponse: tokenResponse,
			JSONResponse:  jsonResponse,
			ExpectedData:  successData,
		},
		{
			Case:                 "Access token from cache",
			AccessToken:          "ACCESSTOKEN",
			AccessTokenFromCache: true,
			JSONResponse:         jsonResponse,
			ExpectedData:         successData,
		},
		{
			Case:                  "Refresh token from cache",
			RefreshToken:          "REFRESH_TOKEN",
			RefreshTokenFromCache: true,
			JSONResponse:          jsonResponse,
			TokenResponse:         tokenResponse,
			ExpectedData:          successData,
		},
		{
			Case:                  "Refresh token from cache, success",
			RefreshToken:          "REFRESH_TOKEN",
			RefreshTokenFromCache: true,
			JSONResponse:          jsonResponse,
			TokenResponse:         tokenResponse,
			ExpectedData:          successData,
		},
		{
			Case:                  "Refresh API error",
			RefreshToken:          "REFRESH_TOKEN",
			RefreshTokenFromCache: true,
			Error:                 fmt.Errorf("API error"),
			ExpectedErrorMessage:  Timeout,
		},
		{
			Case:                  "Refresh API parse error",
			RefreshToken:          "REFRESH_TOKEN",
			RefreshTokenFromCache: true,
			TokenResponse:         "INVALID_JSON",
			ExpectedErrorMessage:  TokenRefreshFailed,
		},
		{
			Case:                 "Default config token",
			RefreshToken:         DefaultRefreshToken,
			ExpectedErrorMessage: InvalidRefreshToken,
		},
		{
			Case:              "Cache data",
			CacheTimeout:      60,
			CacheJSONResponse: jsonResponse,
			ExpectedData:      successData,
		},
		{
			Case:              "Cache data, invalid data",
			CacheTimeout:      60,
			RefreshToken:      "REFRESH_TOKEN",
			CacheJSONResponse: "ERR",
			TokenResponse:     tokenResponse,
			JSONResponse:      jsonResponse,
			ExpectedData:      successData,
		},
		{
			Case:              "Cache data, no cache",
			CacheTimeout:      60,
			RefreshToken:      "REFRESH_TOKEN",
			ResponseCacheMiss: true,
			TokenResponse:     tokenResponse,
			JSONResponse:      jsonResponse,
			ExpectedData:      successData,
		},
		{
			Case:                 "API body failure",
			AccessToken:          "ACCESSTOKEN",
			AccessTokenFromCache: true,
			ResponseCacheMiss:    true,
			JSONResponse:         "ERR",
			ExpectedErrorMessage: "invalid character 'E' looking for beginning of value",
		},
		{
			Case:                 "API request failure",
			AccessToken:          "ACCESSTOKEN",
			AccessTokenFromCache: true,
			ResponseCacheMiss:    true,
			JSONResponse:         "ERR",
			Error:                fmt.Errorf("no response"),
			ExpectedErrorMessage: "no response",
		},
	}

	for _, tc := range cases {
		url := "https://www.strava.com/api/v3/athlete/activities?page=1&per_page=1"
		tokenURL := fmt.Sprintf("https://ohmyposh.dev/api/refresh?segment=test&token=%s", tc.RefreshToken)

		var props properties.Map = map[properties.Property]interface{}{
			properties.CacheTimeout: tc.CacheTimeout,
			properties.AccessToken:  tc.AccessToken,
			properties.RefreshToken: tc.RefreshToken,
		}

		cache := &mock.MockedCache{}

		cache.On("Get", url).Return(tc.CacheJSONResponse, !tc.ResponseCacheMiss)
		cache.On("Get", accessTokenKey).Return(tc.AccessToken, tc.AccessTokenFromCache)
		cache.On("Get", refreshTokenKey).Return(tc.RefreshToken, tc.RefreshTokenFromCache)
		cache.On("Set", mock2.Anything, mock2.Anything, mock2.Anything)

		env := &mock.MockedEnvironment{}

		env.On("Cache").Return(cache)
		env.On("HTTPRequest", url).Return([]byte(tc.JSONResponse), tc.Error)
		env.On("HTTPRequest", tokenURL).Return([]byte(tc.TokenResponse), tc.Error)
		env.On("Error", mock2.Anything).Return()

		oauth := &OAuthRequest{
			AccessTokenKey:  accessTokenKey,
			RefreshTokenKey: refreshTokenKey,
			SegmentName:     "test",
		}
		oauth.Init(env, props)

		got, err := OauthResult[*data](oauth, url, nil)
		assert.Equal(t, tc.ExpectedData, got, tc.Case)
		if len(tc.ExpectedErrorMessage) == 0 {
			assert.Nil(t, err, tc.Case)
		} else {
			assert.Equal(t, tc.ExpectedErrorMessage, err.Error(), tc.Case)
		}
	}
}
