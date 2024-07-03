package http

import (
	"io"
	"net"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/cache/mock"

	"github.com/stretchr/testify/assert"
	testify_ "github.com/stretchr/testify/mock"
)

type MockedEnvironment struct {
	testify_.Mock
}

func (env *MockedEnvironment) Cache() cache.Cache {
	args := env.Called()
	return args.Get(0).(cache.Cache)
}

func (env *MockedEnvironment) HTTPRequest(url string, _ io.Reader, _ int, _ ...RequestModifier) ([]byte, error) {
	args := env.Called(url)
	return args.Get(0).([]byte), args.Error(1)
}

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
		c := &mock.Cache{}

		c.On("Get", url).Return(tc.CacheJSONResponse, !tc.ResponseCacheMiss)
		c.On("Set", testify_.Anything, testify_.Anything, testify_.Anything)

		env := &MockedEnvironment{}

		env.On("Cache").Return(c)
		env.On("HTTPRequest", url).Return([]byte(tc.JSONResponse), tc.Error)

		request := &Request{
			Env:          env,
			CacheTimeout: tc.CacheTimeout,
			HTTPTimeout:  0,
		}

		got, err := Do[*data](request, url, nil)
		assert.Equal(t, tc.ExpectedData, got, tc.Case)
		if len(tc.ExpectedErrorMessage) == 0 {
			assert.Nil(t, err, tc.Case)
		} else {
			assert.Equal(t, tc.ExpectedErrorMessage, err.Error(), tc.Case)
		}
	}
}
