package http

import (
	"net"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"

	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
)

func TestRequestResult(t *testing.T) {
	successData := &data{Hello: "world"}
	jsonResponse := `{ "hello":"world" }`
	url := "https://google.com?q=hello"

	cases := []struct {
		Case string
		// API response
		JSONResponse string
		// Cache
		CacheJSONResponse string
		CacheTimeout      int
		ResponseCacheMiss bool
		// Errors
		Error error
		// Validations
		ExpectedErrorMessage string
		ExpectedData         *data
	}{
		{
			Case:         "No cache",
			JSONResponse: jsonResponse,
			ExpectedData: successData,
		},
		{
			Case:              "Cache",
			CacheJSONResponse: `{ "hello":"mom" }`,
			ExpectedData:      &data{Hello: "mom"},
			CacheTimeout:      10,
		},
		{
			Case:              "Cache miss",
			ResponseCacheMiss: true,
			JSONResponse:      jsonResponse,
			CacheJSONResponse: `{ "hello":"mom" }`,
			ExpectedData:      successData,
			CacheTimeout:      10,
		},
		{
			Case:                 "DNS error",
			Error:                &net.DNSError{IsNotFound: true},
			ExpectedErrorMessage: "lookup : ",
		},
		{
			Case:                 "Response incorrect",
			JSONResponse:         `[`,
			ExpectedErrorMessage: "unexpected end of JSON input",
		},
	}

	for _, tc := range cases {
		var props properties.Map = map[properties.Property]interface{}{
			properties.CacheTimeout: tc.CacheTimeout,
		}

		cache := &mock.MockedCache{}

		cache.On("Get", url).Return(tc.CacheJSONResponse, !tc.ResponseCacheMiss)
		cache.On("Set", mock2.Anything, mock2.Anything, mock2.Anything)

		env := &mock.MockedEnvironment{}

		env.On("Cache").Return(cache)
		env.On("HTTPRequest", url).Return([]byte(tc.JSONResponse), tc.Error)
		env.On("Error", mock2.Anything).Return()

		request := &Request{}
		request.Init(env, props)

		got, err := Do[*data](request, url, nil)
		assert.Equal(t, tc.ExpectedData, got, tc.Case)
		if len(tc.ExpectedErrorMessage) == 0 {
			assert.Nil(t, err, tc.Case)
		} else {
			assert.Equal(t, tc.ExpectedErrorMessage, err.Error(), tc.Case)
		}
	}
}
