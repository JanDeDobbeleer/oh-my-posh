package main

import (
	"errors"
	"fmt"
	"oh-my-posh/mock"
	"oh-my-posh/properties"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStravaSegment(t *testing.T) {
	h, _ := time.ParseDuration("6h")
	sixHoursAgo := time.Now().Add(-h).Format(time.RFC3339)
	h, _ = time.ParseDuration("100h")
	fourDaysAgo := time.Now().Add(-h).Format(time.RFC3339)

	cases := []struct {
		Case                       string
		JSONResponse               string
		AccessToken                string
		RefreshToken               string
		AccessTokenCacheFoundFail  bool
		RefreshTokenCacheFoundFail bool
		InitialAccessToken         string
		InitialRefreshToken        string
		TokenRefreshToken          string
		TokenResponse              string
		TokenTest                  bool
		ExpectedString             string
		ExpectedEnabled            bool
		CacheTimeout               int
		CacheFoundFail             bool
		Template                   string
		Error                      error
		AuthDebugMsg               string
	}{
		{
			Case:                       "No initial tokens",
			InitialAccessToken:         "",
			AccessTokenCacheFoundFail:  true,
			RefreshTokenCacheFoundFail: true,
			TokenTest:                  true,
			AuthDebugMsg:               "invalid refresh token",
		},
		{
			Case:                       "Use initial tokens",
			AccessToken:                "NEW_ACCESSTOKEN",
			InitialAccessToken:         "INITIAL ACCESSTOKEN",
			InitialRefreshToken:        "INITIAL REFRESHTOKEN",
			TokenRefreshToken:          "INITIAL REFRESHTOKEN",
			TokenResponse:              `{ "access_token":"NEW_ACCESSTOKEN","refresh_token":"NEW_REFRESHTOKEN", "expires_in":1234 }`,
			AccessTokenCacheFoundFail:  true,
			RefreshTokenCacheFoundFail: true,
			TokenTest:                  true,
		},
		{
			Case:        "Access token from cache",
			AccessToken: "ACCESSTOKEN",
			TokenTest:   true,
		},
		{
			Case:                       "Refresh token from cache",
			AccessTokenCacheFoundFail:  true,
			RefreshTokenCacheFoundFail: false,
			RefreshToken:               "REFRESHTOKEN",
			TokenRefreshToken:          "REFRESHTOKEN",
			TokenTest:                  true,
			AuthDebugMsg:               "invalid refresh token",
		},
		{
			Case: "Ride 6",
			JSONResponse: `
			[{"type":"Ride","start_date":"` + sixHoursAgo + `","name":"Sesongens første på tjukkas","distance":16144.0}]`,
			Template:        "{{.Ago}} {{.Icon}}",
			ExpectedString:  "6h \uf5a2",
			ExpectedEnabled: true,
		},
		{
			Case: "Run 100",
			JSONResponse: `
			[{"type":"Run","start_date":"` + fourDaysAgo + `","name":"Sesongens første på tjukkas","distance":16144.0,"moving_time":7665}]`,
			Template:        "{{.Ago}} {{.Icon}}",
			ExpectedString:  "4d \ufc0c",
			ExpectedEnabled: true,
		},
		{
			Case:            "Error in retrieving data",
			JSONResponse:    "nonsense",
			Error:           errors.New("Something went wrong"),
			ExpectedEnabled: false,
		},
		{
			Case:            "Empty array",
			JSONResponse:    "[]",
			ExpectedEnabled: false,
		},
		{
			Case: "Run from cache",
			JSONResponse: `
			[{"type":"Run","start_date":"` + fourDaysAgo + `","name":"Sesongens første på tjukkas","distance":16144.0,"moving_time":7665}]`,
			Template:        "{{.Ago}} {{.Icon}}",
			ExpectedString:  "4d \ufc0c",
			ExpectedEnabled: true,
			CacheTimeout:    10,
		},
		{
			Case: "Run from not found cache",
			JSONResponse: `
			[{"type":"Run","start_date":"` + fourDaysAgo + `","name":"Morning ride","distance":16144.0,"moving_time":7665}]`,
			Template:        "{{.Ago}} {{.Icon}} {{.Name}} {{.Hours}}h ago",
			ExpectedString:  "4d \ufc0c Morning ride 100h ago",
			ExpectedEnabled: true,
			CacheTimeout:    10,
			CacheFoundFail:  true,
		},
		{
			Case: "Error parsing response",
			JSONResponse: `
			4tffgt4e4567`,
			Template:        "{{.Ago}}{{.Icon}}",
			ExpectedString:  "50",
			ExpectedEnabled: false,
			CacheTimeout:    10,
		},
		{
			Case: "Faulty template",
			JSONResponse: `
			[{"sgv":50,"direction":"DoubleDown"}]`,
			Template:        "{{.Ago}}{{.Burp}}",
			ExpectedString:  incorrectTemplate,
			ExpectedEnabled: true,
			CacheTimeout:    10,
		},
	}

	for _, tc := range cases {
		env := &mock.MockedEnvironment{}
		url := "https://www.strava.com/api/v3/athlete/activities?page=1&per_page=1"
		tokenURL := fmt.Sprintf("https://ohmyposh.dev/api/refresh?segment=strava&token=%s", tc.TokenRefreshToken)
		var props properties.Map = map[properties.Property]interface{}{
			CacheTimeout: tc.CacheTimeout,
		}
		cache := &mock.MockedCache{}
		cache.On("Get", url).Return(tc.JSONResponse, !tc.CacheFoundFail)
		cache.On("Set", url, tc.JSONResponse, tc.CacheTimeout).Return()

		cache.On("Get", StravaAccessToken).Return(tc.AccessToken, !tc.AccessTokenCacheFoundFail)
		cache.On("Get", StravaRefreshToken).Return(tc.RefreshToken, !tc.RefreshTokenCacheFoundFail)

		cache.On("Set", StravaRefreshToken, "NEW_REFRESHTOKEN", 2*525960)
		cache.On("Set", StravaAccessToken, "NEW_ACCESSTOKEN", 20)

		env.On("HTTPRequest", url).Return([]byte(tc.JSONResponse), tc.Error)
		env.On("HTTPRequest", tokenURL).Return([]byte(tc.TokenResponse), tc.Error)
		env.On("Cache").Return(cache)

		if tc.InitialAccessToken != "" {
			props[properties.AccessToken] = tc.InitialAccessToken
		}
		if tc.InitialRefreshToken != "" {
			props[properties.RefreshToken] = tc.InitialRefreshToken
		}

		ns := &Strava{
			props: props,
			env:   env,
		}

		if tc.TokenTest {
			// continue
			at, err := ns.getAccessToken()
			if err != nil {
				if authErr, ok := err.(*AuthError); ok {
					assert.Equal(t, tc.AuthDebugMsg, authErr.Error(), tc.Case)
				} else {
					assert.Equal(t, tc.Error, err, tc.Case)
				}
			} else {
				assert.Equal(t, tc.AccessToken, at, tc.Case)
			}
			continue
		}

		enabled := ns.Enabled()
		assert.Equal(t, tc.ExpectedEnabled, enabled, tc.Case)
		if !enabled {
			continue
		}

		if tc.Template == "" {
			tc.Template = ns.Template()
		}
		var got = renderTemplate(env, tc.Template, ns)

		assert.Equal(t, tc.ExpectedString, got, tc.Case)
	}
}
