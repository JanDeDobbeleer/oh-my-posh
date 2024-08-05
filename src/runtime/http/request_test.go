package http

import (
	"io"
	"net"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"

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
		Error                error
		ExpectedData         *data
		Case                 string
		JSONResponse         string
		CacheJSONResponse    string
		ExpectedErrorMessage string
		CacheTimeout         int
		ResponseCacheMiss    bool
	}{
		{
			Case:         "No cache",
			JSONResponse: jsonResponse,
			ExpectedData: successData,
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
		env := &MockedEnvironment{}
		env.On("HTTPRequest", url).Return([]byte(tc.JSONResponse), tc.Error)

		request := &Request{
			Env:         env,
			HTTPTimeout: 0,
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
