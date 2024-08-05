package http

import (
	"fmt"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache/mock"
	"github.com/stretchr/testify/assert"
	testify_ "github.com/stretchr/testify/mock"
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
		Error                 error
		ExpectedData          *data
		AccessToken           string
		RefreshToken          string
		TokenResponse         string
		JSONResponse          string
		CacheJSONResponse     string
		Case                  string
		ExpectedErrorMessage  string
		CacheTimeout          int
		ResponseCacheMiss     bool
		AccessTokenFromCache  bool
		RefreshTokenFromCache bool
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
			Case:          "Cache data, invalid data",
			RefreshToken:  "REFRESH_TOKEN",
			TokenResponse: tokenResponse,
			JSONResponse:  jsonResponse,
			ExpectedData:  successData,
		},
		{
			Case:          "Cache data, no cache",
			RefreshToken:  "REFRESH_TOKEN",
			TokenResponse: tokenResponse,
			JSONResponse:  jsonResponse,
			ExpectedData:  successData,
		},
		{
			Case:                 "API body failure",
			AccessToken:          "ACCESSTOKEN",
			AccessTokenFromCache: true,
			JSONResponse:         "ERR",
			ExpectedErrorMessage: "invalid character 'E' looking for beginning of value",
		},
		{
			Case:                 "API request failure",
			AccessToken:          "ACCESSTOKEN",
			AccessTokenFromCache: true,
			JSONResponse:         "ERR",
			Error:                fmt.Errorf("no response"),
			ExpectedErrorMessage: "no response",
		},
	}

	for _, tc := range cases {
		url := "https://www.strava.com/api/v3/athlete/activities?page=1&per_page=1"
		tokenURL := fmt.Sprintf("https://ohmyposh.dev/api/refresh?segment=test&token=%s", tc.RefreshToken)

		cache := &mock.Cache{}

		cache.On("Get", accessTokenKey).Return(tc.AccessToken, tc.AccessTokenFromCache)
		cache.On("Get", refreshTokenKey).Return(tc.RefreshToken, tc.RefreshTokenFromCache)
		cache.On("Set", testify_.Anything, testify_.Anything, testify_.Anything)

		env := &MockedEnvironment{}

		env.On("Cache").Return(cache)
		env.On("HTTPRequest", url).Return([]byte(tc.JSONResponse), tc.Error)
		env.On("HTTPRequest", tokenURL).Return([]byte(tc.TokenResponse), tc.Error)
		env.On("Error", testify_.Anything)

		oauth := &OAuthRequest{
			AccessTokenKey:  accessTokenKey,
			RefreshTokenKey: refreshTokenKey,
			SegmentName:     "test",
			AccessToken:     tc.AccessToken,
			RefreshToken:    tc.RefreshToken,
			Request: Request{
				Env:         env,
				HTTPTimeout: 20,
			},
		}

		got, err := OauthResult[*data](oauth, url, nil)
		assert.Equal(t, tc.ExpectedData, got, tc.Case)

		if len(tc.ExpectedErrorMessage) == 0 {
			assert.Nil(t, err, tc.Case)
		} else {
			assert.Equal(t, tc.ExpectedErrorMessage, err.Error(), tc.Case)
		}
	}
}
