package segments

import (
	"path"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"

	"github.com/stretchr/testify/assert"
)

func TestGcpSegment(t *testing.T) {
	cases := []struct {
		Case            string
		CfgData         string
		ActiveConfig    string
		EnvActiveConfig string
		ExpectedString  string
		ExpectedEnabled bool
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
		{
			Case:            "use CLOUDSDK_ACTIVE_CONFIG_NAME",
			EnvActiveConfig: "myconfig",
			ExpectedEnabled: true,
			CfgData: `
			[core]
			account = user@example.com
			project = cloud-proj

			[compute]
			region = us-west1
			`,
			ExpectedString: "cloud-proj :: us-west1 :: user@example.com",
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("Getenv", "CLOUDSDK_CONFIG").Return("config")
		env.On("Getenv", "CLOUDSDK_ACTIVE_CONFIG_NAME").Return(tc.EnvActiveConfig)

		// Only use fallback file if env var is not set
		if tc.EnvActiveConfig == "" {
			fcPath := path.Join("config", "active_config")
			env.On("FileContent", fcPath).Return(tc.ActiveConfig)
		}

		// Resolve active config name
		activeConfig := tc.EnvActiveConfig
		if activeConfig == "" {
			activeConfig = tc.ActiveConfig
		}

		cfgpath := path.Join("config", "configurations", "config_"+activeConfig)
		env.On("FileContent", cfgpath).Return(tc.CfgData)

		g := &Gcp{}
		g.Init(options.Map{}, env)

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
			GOOS:     runtime.WINDOWS,
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
		env := new(mock.Environment)
		env.On("Getenv", "CLOUDSDK_CONFIG").Return(tc.CloudSDKConfig)
		env.On("Getenv", "APPDATA").Return(tc.AppData)
		env.On("Home").Return(tc.Home)
		env.On("GOOS").Return(tc.GOOS)

		g := &Gcp{}
		g.Init(options.Map{}, env)

		assert.Equal(t, tc.Expected, g.getConfigDirectory(), tc.Case)
	}
}

func TestGetActiveConfig(t *testing.T) {
	cases := []struct {
		Case                    string
		EnvActiveConfigName     string
		FileActiveConfigContent string
		ExpectedString          string
		ExpectedError           string
	}{
		{
			Case:                "CLOUDSDK_ACTIVE_CONFIG_NAME set",
			EnvActiveConfigName: "envconfig",
			ExpectedString:      "envconfig",
		},
		{
			Case:                    "Fallback to file content",
			FileActiveConfigContent: "fileconfig",
			ExpectedString:          "fileconfig",
		},
		{
			Case:          "No config anywhere",
			ExpectedError: GCPNOACTIVECONFIG,
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("Getenv", "CLOUDSDK_ACTIVE_CONFIG_NAME").Return(tc.EnvActiveConfigName)

		// If env var not set, mock file fallback
		if tc.EnvActiveConfigName == "" {
			env.On("FileContent", path.Join("", "active_config")).Return(tc.FileActiveConfigContent)
		}

		g := &Gcp{}
		g.Init(options.Map{}, env)

		got, err := g.getActiveConfig("")
		assert.Equal(t, tc.ExpectedString, got, tc.Case)
		if len(tc.ExpectedError) > 0 {
			assert.EqualError(t, err, tc.ExpectedError, tc.Case)
		} else {
			assert.NoError(t, err, tc.Case)
		}
	}
}
