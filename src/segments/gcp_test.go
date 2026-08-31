package segments

import (
	"errors"
	"path"
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/segments/options"
	"github.com/jandedobbeleer/oh-my-posh/src/template"

	"github.com/stretchr/testify/assert"
)

func TestGcpSegment(t *testing.T) {
	cases := []struct {
		ConfigHelperError   error
		ExpectedAuthError   string
		CfgData             string
		ActiveConfig        string
		EnvActiveConfig     string
		Case                string
		RenderTemplate      string
		ConfigHelperOutput  string
		ExpectedString      string
		ExpectedTokenExpiry string
		ExpectedAuthStatus  string
		ReferencedFields    []string
		ExpectedAuthorized  bool
		ExpectedEnabled     bool
		ExpectAuthProbe     bool
		HasGcloud           bool
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
			Case:            "does not fetch auth without auth references",
			ExpectedEnabled: true,
			ActiveConfig:    "production",
			CfgData: `
			[core]
			account = test@example.com
			project = test-test-test

			[compute]
			region = europe-test1
			`,
			ReferencedFields:   []string{"Project"},
			ExpectedAuthStatus: "",
		},
		{
			Case:            "fetches authorized status when referenced",
			ExpectedEnabled: true,
			ActiveConfig:    "production",
			CfgData: `
			[core]
			account = test@example.com
			project = test-test-test

			[compute]
			region = europe-test1
			`,
			ReferencedFields:    []string{"Authorized"},
			RenderTemplate:      "{{.Authorized}} :: {{.AuthStatus}} :: {{.TokenExpiry}}",
			ExpectAuthProbe:     true,
			HasGcloud:           true,
			ConfigHelperOutput:  `{"credential":{"access_token":"token","token_expiry":"2026-08-31T16:00:00Z"}}`,
			ExpectedString:      "true :: authorized :: 2026-08-31T16:00:00Z",
			ExpectedAuthStatus:  "authorized",
			ExpectedAuthorized:  true,
			ExpectedTokenExpiry: "2026-08-31T16:00:00Z",
		},
		{
			Case:            "fetches unauthorized status when config helper fails",
			ExpectedEnabled: true,
			ActiveConfig:    "production",
			CfgData: `
			[core]
			account = test@example.com
			project = test-test-test

			[compute]
			region = europe-test1
			`,
			ReferencedFields:   []string{"AuthStatus"},
			ExpectAuthProbe:    true,
			HasGcloud:          true,
			ConfigHelperError:  errors.New("reauthentication required"),
			ExpectedAuthStatus: "unauthorized",
			ExpectedAuthError:  "reauthentication required",
		},
		{
			Case:            "fetches unknown status when gcloud is unavailable",
			ExpectedEnabled: true,
			ActiveConfig:    "production",
			CfgData: `
			[core]
			account = test@example.com
			project = test-test-test

			[compute]
			region = europe-test1
			`,
			ReferencedFields:   []string{"AuthStatus"},
			ExpectAuthProbe:    true,
			ExpectedAuthStatus: "unknown",
		},
		{
			Case:            "fetches unauthorized status when no access token is returned",
			ExpectedEnabled: true,
			ActiveConfig:    "production",
			CfgData: `
			[core]
			account = test@example.com
			project = test-test-test

			[compute]
			region = europe-test1
			`,
			ReferencedFields:   []string{"TokenExpiry"},
			ExpectAuthProbe:    true,
			HasGcloud:          true,
			ConfigHelperOutput: `{"credential":{}}`,
			ExpectedAuthStatus: "unauthorized",
		},
		{
			Case:            "fetches unauthorized status when config helper returns empty output",
			ExpectedEnabled: true,
			ActiveConfig:    "production",
			CfgData: `
			[core]
			account = test@example.com
			project = test-test-test

			[compute]
			region = europe-test1
			`,
			ReferencedFields:   []string{"AuthError"},
			ExpectAuthProbe:    true,
			HasGcloud:          true,
			ExpectedAuthStatus: "unauthorized",
			ExpectedAuthError:  "empty auth response",
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
		if len(tc.ReferencedFields) != 0 {
			g.SetReferencedFields(template.RefSet{Fields: tc.ReferencedFields, Analyzable: true})
		}
		if tc.ExpectAuthProbe {
			env.On("HasCommand", gcpCommand).Return(tc.HasGcloud)
		}
		if tc.HasGcloud {
			env.On("RunCommand", gcpCommand, []string{"config", "config-helper", "--format=json"}).
				Return(tc.ConfigHelperOutput, tc.ConfigHelperError)
		}

		assert.Equal(t, tc.ExpectedEnabled, g.Enabled(), tc.Case)
		assert.Equal(t, tc.ExpectedAuthStatus, g.AuthStatus, tc.Case)
		assert.Equal(t, tc.ExpectedAuthError, g.AuthError, tc.Case)
		assert.Equal(t, tc.ExpectedAuthorized, g.Authorized, tc.Case)
		assert.Equal(t, tc.ExpectedTokenExpiry, g.TokenExpiry, tc.Case)
		if !tc.ExpectAuthProbe {
			env.AssertNotCalled(t, "HasCommand", gcpCommand)
		}

		if !tc.ExpectedEnabled || tc.ExpectedString == "" {
			continue
		}

		renderTemplateText := tc.RenderTemplate
		if renderTemplateText == "" {
			renderTemplateText = "{{.Project}} :: {{.Region}} :: {{.Account}}"
		}

		assert.Equal(t, tc.ExpectedString, renderTemplate(env, renderTemplateText, g), tc.Case)
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
