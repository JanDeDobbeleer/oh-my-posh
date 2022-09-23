package segments

import (
	"path"
	"testing"

	"oh-my-posh/environment"
	"oh-my-posh/mock"

	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
)

func TestGcpSegment(t *testing.T) {
	cases := []struct {
		Case            string
		CfgData         string
		ActiveConfig    string
		ExpectedEnabled bool
		ExpectedString  string
	}{
		{
			Case:            "happy path",
			ExpectedEnabled: true,
			ActiveConfig:    "production",
			CfgData: `
			[core]
			account = test@example.com
			project = test-test-test

			[compute]
			region = europe-test1
			`,
			ExpectedString: "test-test-test :: europe-test1 :: test@example.com",
		},
		{
			Case:            "no active config",
			ExpectedEnabled: false,
		},
		{
			Case:            "empty config",
			ActiveConfig:    "production",
			ExpectedEnabled: false,
		},
		{
			Case:            "bad config",
			ActiveConfig:    "production",
			CfgData:         "{bad}",
			ExpectedEnabled: false,
		},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("Getenv", "CLOUDSDK_CONFIG").Return("config")
		fcPath := path.Join("config", "active_config")
		env.On("FileContent", fcPath).Return(tc.ActiveConfig)
		cfgpath := path.Join("config", "configurations", "config_production")
		env.On("FileContent", cfgpath).Return(tc.CfgData)
		env.On("Log", environment.Error, "Gcp.Enabled()", mock2.Anything).Return()
		g := &Gcp{
			env: env,
		}
		assert.Equal(t, tc.ExpectedEnabled, g.Enabled(), tc.Case)
		if tc.ExpectedEnabled {
			assert.Equal(t, tc.ExpectedString, renderTemplate(env, "{{.Project}} :: {{.Region}} :: {{.Account}}", g), tc.Case)
		}
	}
}

func TestGetConfigDirectory(t *testing.T) {
	cases := []struct {
		Case           string
		GOOS           string
		Home           string
		AppData        string
		CloudSDKConfig string
		Expected       string
	}{
		{
			Case:           "CLOUDSDK_CONFIG",
			CloudSDKConfig: "/Users/posh/.config/gcloud",
			Expected:       "/Users/posh/.config/gcloud",
		},
		{
			Case:     "Windows",
			GOOS:     environment.WINDOWS,
			AppData:  "/Users/posh/.config",
			Expected: "/Users/posh/.config/gcloud",
		},
		{
			Case:     "default",
			Home:     "/Users/posh2/",
			Expected: "/Users/posh2/.config/gcloud",
		},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("Getenv", "CLOUDSDK_CONFIG").Return(tc.CloudSDKConfig)
		env.On("Getenv", "APPDATA").Return(tc.AppData)
		env.On("Home").Return(tc.Home)
		env.On("GOOS").Return(tc.GOOS)
		g := &Gcp{
			env: env,
		}
		assert.Equal(t, tc.Expected, g.getConfigDirectory(), tc.Case)
	}
}

func TestGetActiveConfig(t *testing.T) {
	cases := []struct {
		Case           string
		ActiveConfig   string
		ExpectedString string
		ExpectedError  string
	}{
		{
			Case:          "No active config",
			ExpectedError: GCPNOACTIVECONFIG,
		},
		{
			Case:           "No active config",
			ActiveConfig:   "production",
			ExpectedString: "production",
		},
	}

	for _, tc := range cases {
		env := new(mock.MockedEnvironment)
		env.On("FileContent", "active_config").Return(tc.ActiveConfig)
		g := &Gcp{
			env: env,
		}
		got, err := g.getActiveConfig("")
		assert.Equal(t, tc.ExpectedString, got, tc.Case)
		if len(tc.ExpectedError) > 0 {
			assert.EqualError(t, err, tc.ExpectedError, tc.Case)
		} else {
			assert.NoError(t, err, tc.Case)
		}
	}
}
