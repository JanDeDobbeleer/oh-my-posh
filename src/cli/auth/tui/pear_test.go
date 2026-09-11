package tui

import (
	"errors"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/cli/auth"
	runtime_ "github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPear_Authenticate(t *testing.T) {
	testCases := []struct {
		name           string
		response       string
		requestError   error
		expectedError  error
		expectedToken  string
		shouldSetToken bool
	}{
		{
			name:           "successful authentication",
			response:       `{"accessToken":"test-token-123"}`,
			expectedToken:  "test-token-123",
			shouldSetToken: true,
		},
		{
			name:          "request fails",
			requestError:  errors.New("failed to request token"),
			expectedError: errors.New("failed to request token"),
		},
		{
			name:          "invalid response JSON",
			response:      `{"invalid":"json"}`,
			expectedError: errors.New("received empty access token"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			env := &runtime_.Environment{}

			env.On("HTTPRequest", pearAuthURL).Return([]byte(tc.response), tc.requestError)

			pear := &Pear{
				env: env,
			}

			pear.Authenticate()

			if tc.expectedError != nil {
				require.NotNil(t, pear.err)
				assert.Equal(t, tc.expectedError.Error(), pear.err.Error())
			} else {
				assert.Nil(t, pear.err)
			}

			if tc.shouldSetToken {
				token, ok := cache.Device.Get[string](auth.PEARTOKEN)
				require.True(t, ok)
				assert.Equal(t, tc.expectedToken, token)
			}

			cache.Device.DeleteAll()
		})
	}
}
